package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaskEmail(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "return empty string on empty email address",
			args: args{""},
			want: "",
		},
		{
			name: "empty name part",
			args: args{"@domain.com"},
			want: "******@domain.com",
		},
		{
			name: "mask start reduced and padding applied when name length < 6",
			args: args{"123@domain.com"},
			want: "1*****@domain.com",
		},
		{
			name: "start mask at index 4 and mask everything until '@' rune",
			args: args{"really_long_test_email_help_when_does_it_stop@domain.com"},
			want: "rea******************************************@domain.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MaskEmail(tt.args.email), "MaskEmail(%v)", tt.args.email)
		})
	}
}

func TestMaskUsername(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "return empty string on empty username",
			args: args{username: ""},
			want: "",
		},
		{
			name: "mask everything if username length == 1",
			args: args{username: "X"},
			want: "*",
		},
		{
			name: "mask and pad when username length 2 or 3",
			args: args{username: "xx"},
			want: "x****x",
		},
		{
			name: "mask everything but first and last rune when username size > 3",
			args: args{username: "test_username"},
			want: "t***********e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MaskUsername(tt.args.username), "MaskUsername(%v)", tt.args.username)
		})
	}
}
