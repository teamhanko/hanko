package user

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/dto"
)

// ======================================================
// MODELS
// ======================================================

type FirebaseUser struct {
	LocalID       string `json:"localId"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	PasswordHash  string `json:"passwordHash"`
	Salt          string `json:"salt"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"lastLoginAt"`
}

type FirebaseHashConfig struct {
	Algorithm           string `json:"algorithm"`
	Base64SignerKey     string `json:"base64_signer_key"`
	Base64SaltSeparator string `json:"base64_salt_separator"`
	Rounds              int    `json:"rounds"`
	MemCost             int    `json:"mem_cost"`
}

func (f *FirebaseHashConfig) Validate() error {
	if f.Algorithm != "SCRYPT" {
		return errors.New("crypto: invalid algorithm")
	}

	if f.Base64SignerKey == "" {
		return errors.New("crypto: invalid base64_signer_key")
	}

	if f.Base64SaltSeparator == "" {
		return errors.New("crypto: invalid base64_salt_separator")
	}

	// Firebase computes N as 2^memCost; ensure N fits in int, so we can fail early
	if f.MemCost <= 0 || f.MemCost >= 63 {
		return errors.New("crypto: invalid memCost, must be between 1 and 62")
	}

	n64 := uint64(1) << f.MemCost

	if n64 > uint64(math.MaxInt) {
		return errors.New("crypto: memCost (N) does not fit in int")
	}

	if f.Rounds <= 0 {
		return errors.New("crypto: invalid rounds (scrypt 'r') <= 0")
	}

	return nil
}

type result struct {
	entry ImportOrExportEntry
	raw   FirebaseUser
	err   error
}

type dlqEntry struct {
	Error string       `json:"error"`
	User  FirebaseUser `json:"user"`
}

type writerReport struct {
	Written int
	Failed  int
}

// ======================================================
// CLI
// ======================================================

type options struct {
	inputFile  string
	configFile string
	outputFile string
	dlqFile    string
	workers    int
}

func NewFirebaseCommand() *cobra.Command {
	opts := &options{
		outputFile: "hanko-import.json",
		dlqFile:    "firebase-dlq.jsonl",
		workers:    runtime.NumCPU(),
	}

	cmd := &cobra.Command{
		Use:   "firebase",
		Short: "Convert user data exported via Firebase CLI to Hanko import data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.inputFile, "input", "", "Firebase export input file (JSON)")
	cmd.Flags().StringVar(&opts.configFile, "config", "", "Firebase hash config file (JSON)")
	cmd.Flags().StringVar(&opts.outputFile, "output", opts.outputFile, "Hanko import output file (JSON)")
	cmd.Flags().StringVar(&opts.dlqFile, "dlq", opts.dlqFile, "DLQ file containing user conversion errors (NDJSON)")
	cmd.Flags().IntVar(&opts.workers, "workers", opts.workers, "Number of workers")

	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("config")

	return cmd
}

var v = validator.New()

// ======================================================
// MAIN
// ======================================================

func run(opts *options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := loadFirebaseConfig(opts.configFile)
	if err != nil {
		return fmt.Errorf("firebase hash config: %w", err)
	}

	inFile, err := os.Open(opts.inputFile)
	if err != nil {
		return fmt.Errorf("firebase user input: %w", err)
	}
	defer inFile.Close()

	tmpFile, err := os.CreateTemp("", "firebase-*.ndjson")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tmpWriter := bufio.NewWriter(tmpFile)

	dlqFile, err := os.Create(opts.dlqFile)
	if err != nil {
		return fmt.Errorf("dlq: %w", err)
	}
	defer dlqFile.Close()

	dlqWriter := bufio.NewWriter(dlqFile)

	jobs := make(chan FirebaseUser, opts.workers*2)
	results := make(chan result, opts.workers*2)

	// -----------------------------
	// WORKERS
	// -----------------------------
	var wg sync.WaitGroup

	for i := 0; i < opts.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, cfg, jobs, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// -----------------------------
	// STREAM INPUT
	// -----------------------------
	streamErr := make(chan error, 1)

	go func() {
		defer close(jobs)
		streamErr <- stream(ctx, inFile, jobs)
	}()

	if err := <-streamErr; err != nil {
		cancel()
		return fmt.Errorf("stream: %w", err)
	}

	// -----------------------------
	// WRITER
	// -----------------------------
	report, err := writer(
		ctx,
		tmpWriter,
		dlqWriter,
		results,
	)
	if err != nil {
		return fmt.Errorf("writer failed: %w", err)
	}

	if report.Written == 0 {
		return fmt.Errorf("could not convert any users, no output file written, number of failed entries that went to DLQ=%d)", report.Failed)
	}

	// -----------------------------
	// FINALIZE JSON ARRAY
	// -----------------------------

	if err := tmpWriter.Flush(); err != nil {
		return err
	}
	if err := dlqWriter.Flush(); err != nil {
		return err
	}

	outFile, err := os.Create(opts.outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	outWriter := bufio.NewWriter(outFile)

	if _, err := outWriter.WriteString("[\n"); err != nil {
		return err
	}

	// rewind temp file
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(tmpFile)

	first := true
	for scanner.Scan() {
		if !first {
			outWriter.WriteString(",\n")
		}
		first = false

		outWriter.Write(scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if _, err := outWriter.WriteString("\n]\n"); err != nil {
		return err
	}

	if err := outWriter.Flush(); err != nil {
		return err
	}

	if report.Failed > 0 {
		log.Printf("conversion completed with partial failures: written=%d failed=%d\n", report.Written, report.Failed)
	} else {
		log.Printf("conversion completed: written=%d failed=%d\n", report.Written, report.Failed)
	}

	return nil
}

// ======================================================
// STREAMING INPUT
// ======================================================

func stream(ctx context.Context, in *os.File, jobs chan<- FirebaseUser) error {
	dec := json.NewDecoder(bufio.NewReader(in))

	_, err := dec.Token() // {
	if err != nil {
		return err
	}

	for dec.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		t, err := dec.Token()
		if err != nil {
			return err
		}

		if t.(string) != "users" {
			var discard any
			_ = dec.Decode(&discard)
			continue
		}

		_, err = dec.Token() // [
		if err != nil {
			return err
		}

		for dec.More() {
			var u FirebaseUser
			if err := dec.Decode(&u); err != nil {
				return err
			}

			jobs <- u
		}

		_, err = dec.Token() // ]
		if err != nil {
			return err
		}
	}

	_, err = dec.Token() // }
	return err
}

// ======================================================
// WORKER
// ======================================================

func worker(ctx context.Context, cfg FirebaseHashConfig, jobs <-chan FirebaseUser, results chan<- result) {
	for {
		select {
		case <-ctx.Done():
			return

		case u, ok := <-jobs:
			if !ok {
				return
			}

			entry, err := convertUser(u, cfg)

			select {
			case results <- result{entry: entry, raw: u, err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// ======================================================
// WRITER
// ======================================================

func writer(
	ctx context.Context,
	w *bufio.Writer,
	dlq *bufio.Writer,
	results <-chan result,
) (writerReport, error) {

	var written, failed int

	for {
		select {

		case <-ctx.Done():
			return writerReport{written, failed},
				ctx.Err()

		case r, ok := <-results:
			if !ok {
				if err := w.Flush(); err != nil {
					return writerReport{written, failed}, err
				}
				if err := dlq.Flush(); err != nil {
					return writerReport{written, failed}, err
				}

				return writerReport{written, failed}, nil
			}

			// -------------------------
			// DLQ FOR FAILED ENTRIES
			// -------------------------
			if r.err != nil {
				failed++

				b, err := json.Marshal(dlqEntry{
					User:  r.raw,
					Error: r.err.Error(),
				})
				if err != nil {
					return writerReport{written, failed}, err
				}

				if _, err := dlq.Write(b); err != nil {
					return writerReport{written, failed}, err
				}
				dlq.WriteByte('\n')
				continue
			}

			// -------------------------
			// SUCCESS → NDJSON
			// -------------------------
			written++

			b, err := json.Marshal(r.entry)
			if err != nil {
				return writerReport{written, failed}, err
			}

			if _, err := w.Write(b); err != nil {
				return writerReport{written, failed}, err
			}
			w.WriteByte('\n')
		}
	}
}

// ======================================================
// TRANSFORM
// ======================================================

func convertUser(u FirebaseUser, cfg FirebaseHashConfig) (ImportOrExportEntry, error) {

	if u.Email == "" && u.PasswordHash == "" {
		// If both are missing, there is no way to authenticate the user, unless some other credential/factor
		// is manually added after the firebase -> hanko user conversion. This feels unlikely, so we count this
		// as a failure.
		return ImportOrExportEntry{}, fmt.Errorf("email and passwordHash missing for Firebase user with localId '%s'", u.LocalID)
	}

	fscryptString, err := buildFbscryptString(u, cfg)
	if err != nil {
		return ImportOrExportEntry{}, err
	}

	convertedEntry := ImportOrExportEntry{}

	if u.Email != "" {
		convertedEntry.Emails = []ImportOrExportEmail{
			{
				Address:    u.Email,
				IsPrimary:  true,            // we assume imported users are new users, so we set it primary
				IsVerified: u.EmailVerified, // if not returned by the firebase export we accept that it zeros to false
			},
		}
	}

	if u.PasswordHash != "" {
		convertedEntry.Password = &ImportPasswordCredential{
			Password: fscryptString,
		}
	} else {
		log.Printf("[WARN] no password given for Firebase user with localId '%s', converting anyway\n", u.LocalID)
	}

	createdAt, _ := parseTime(u.CreatedAt)
	updatedAt, _ := parseTime(u.UpdatedAt)

	convertedEntry.CreatedAt = createdAt
	convertedEntry.UpdatedAt = updatedAt

	err = convertedEntry.validate(v)
	if err != nil {
		vErrs := dto.TransformValidationErrors(err)
		vErr := fmt.Errorf("%v", strings.Join(vErrs, " and "))
		return ImportOrExportEntry{}, vErr
	}

	return convertedEntry, nil
}

// ======================================================
// FIREBASE STRING (PARAMETERS INCL. PASSWORD HASH)
// ======================================================

func buildFbscryptString(u FirebaseUser, cfg FirebaseHashConfig) (string, error) {
	meta := fmt.Sprintf(
		"v=1,n=%d,r=%d,p=%d,ss=%s,sk=%s",
		cfg.MemCost,
		cfg.Rounds,
		1,
		cfg.Base64SaltSeparator,
		cfg.Base64SignerKey,
	)

	return fmt.Sprintf(
		"$fbscrypt$%s$%s$%s",
		meta,
		u.Salt,
		u.PasswordHash,
	), nil
}

// ======================================================
// CONFIG
// ======================================================

func loadFirebaseConfig(path string) (FirebaseHashConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return FirebaseHashConfig{}, err
	}

	var cfg FirebaseHashConfig
	err = json.Unmarshal(b, &cfg)

	if err != nil {
		return FirebaseHashConfig{}, err
	}

	// Validate and possibly fail early since the config is declared as param in multiple functions
	if err = cfg.Validate(); err != nil {
		return FirebaseHashConfig{}, err
	}

	return cfg, err
}

// ======================================================
// HELPERS
// ======================================================

func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	timestampMillis, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}

	return new(time.Unix(0, timestampMillis*int64(time.Millisecond))), nil
}
