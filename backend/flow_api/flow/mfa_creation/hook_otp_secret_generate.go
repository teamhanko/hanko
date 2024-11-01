package mfa_creation

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/pquerna/otp/totp"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"image/png"
)

type OTPSecretGenerate struct {
	shared.Action
}

func (h OTPSecretGenerate) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if c.GetCurrentState() == shared.StateMFAOTPSecretCreation &&
		c.Stash().Get(shared.StashPathOTPSecret).Exists() &&
		c.Stash().Get(shared.StashPathOTPImageSource).Exists() {

		otpSecret := c.Stash().Get(shared.StashPathOTPSecret).String()
		otpImageSource := c.Stash().Get(shared.StashPathOTPImageSource).String()

		_ = c.Payload().Set("otp_secret", otpSecret)
		_ = c.Payload().Set("otp_image_source", otpImageSource)

		return nil
	}

	userEmail := c.Stash().Get(shared.StashPathEmail).String()
	userUsername := c.Stash().Get(shared.StashPathUsername).String()

	if userEmail == "" && userUsername == "" {
		return errors.New("could not create OTP secret: no email or username found on the stash")
	}

	otpAccountName := userEmail
	if userEmail == "" {
		otpAccountName = userUsername
	}

	otpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      deps.Cfg.Service.Name,
		AccountName: otpAccountName,
	})

	if err != nil {
		return fmt.Errorf("could not generate OTP key: %w", err)
	}

	otpImage, err := otpKey.Image(200, 200)
	if err != nil {
		return fmt.Errorf("could not generate OTP image: %w", err)
	}

	otpImagePNGBuffer := new(bytes.Buffer)
	err = png.Encode(otpImagePNGBuffer, otpImage)
	if err != nil {
		return fmt.Errorf("could not PNG encode OTP image: %w", err)
	}

	otpImageSource := fmt.Sprintf(
		"data:image/png;base64,%s", base64.StdEncoding.EncodeToString(otpImagePNGBuffer.Bytes()))

	otpSecret := otpKey.Secret()

	_ = c.Stash().Set(shared.StashPathOTPSecret, otpSecret)
	_ = c.Stash().Set(shared.StashPathOTPImageSource, otpImageSource)

	_ = c.Payload().Set("otp_secret", otpSecret)
	_ = c.Payload().Set("otp_image_source", otpImageSource)

	return nil
}
