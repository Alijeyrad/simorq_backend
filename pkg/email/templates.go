package email

import (
	"fmt"
)

// WaitlistEmailData contains the data needed for waitlist email templates.
type WaitlistEmailData struct {
	FirstName       string
	Email           string
	VerificationURL string
	InvitationURL   string
	Position        int
	ReferralCode    string
	AppName         string
	BaseURL         string
}

// BuildWaitlistVerificationEmail creates a verification email message for new waitlist signups.
func BuildWaitlistVerificationEmail(data WaitlistEmailData) Message {
	appName := data.AppName
	if appName == "" {
		appName = "Simorgh"
	}

	firstName := data.FirstName
	if firstName == "" {
		firstName = "there"
	}

	subject := fmt.Sprintf("Verify your email to join the %s waitlist", appName)

	textBody := fmt.Sprintf(`Hi %s,

Thanks for joining the %s waitlist!

Please verify your email by clicking the link below:
%s

Your current position: #%d

Share your referral code with friends to move up the queue: %s

Thanks,
The %s Team`,
		firstName, appName, data.VerificationURL, data.Position, data.ReferralCode, appName)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #2563eb;">Hi %s,</h2>
    <p>Thanks for joining the %s waitlist!</p>
    <p>Please verify your email by clicking the button below:</p>
    <p style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;">Verify Email</a>
    </p>
    <p>Your current position: <strong>#%d</strong></p>
    <p>Share your referral code with friends to move up the queue:</p>
    <p style="background-color: #f3f4f6; padding: 10px 15px; border-radius: 4px; font-family: monospace; font-size: 16px;">%s</p>
    <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">Thanks,<br>The %s Team</p>
</body>
</html>`,
		firstName, appName, data.VerificationURL, data.Position, data.ReferralCode, appName)

	return Message{
		To:       []string{data.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}

// BuildWaitlistInvitationEmail creates an invitation email for waitlist entries.
func BuildWaitlistInvitationEmail(data WaitlistEmailData) Message {
	appName := data.AppName
	if appName == "" {
		appName = "Simorgh"
	}

	firstName := data.FirstName
	if firstName == "" {
		firstName = "there"
	}

	subject := fmt.Sprintf("You're invited to join %s!", appName)

	textBody := fmt.Sprintf(`Hi %s,

Great news! You've been invited to join %s.

Click the link below to accept your invitation and create your account:
%s

This invitation expires in 30 days.

Thanks,
The %s Team`,
		firstName, appName, data.InvitationURL, appName)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #2563eb;">Hi %s,</h2>
    <p>Great news! You've been invited to join %s.</p>
    <p>Click the button below to accept your invitation and create your account:</p>
    <p style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background-color: #16a34a; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-size: 16px;">Accept Invitation</a>
    </p>
    <p style="color: #6b7280; font-size: 14px;"><em>This invitation expires in 30 days.</em></p>
    <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">Thanks,<br>The %s Team</p>
</body>
</html>`,
		firstName, appName, data.InvitationURL, appName)

	return Message{
		To:       []string{data.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}

// BuildWaitlistResendVerificationEmail creates a resend verification email message.
func BuildWaitlistResendVerificationEmail(data WaitlistEmailData) Message {
	appName := data.AppName
	if appName == "" {
		appName = "Simorgh"
	}

	firstName := data.FirstName
	if firstName == "" {
		firstName = "there"
	}

	subject := fmt.Sprintf("Verify your email for %s", appName)

	textBody := fmt.Sprintf(`Hi %s,

You requested a new verification link for the %s waitlist.

Please verify your email by clicking the link below:
%s

Your current position: #%d

Thanks,
The %s Team`,
		firstName, appName, data.VerificationURL, data.Position, appName)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #2563eb;">Hi %s,</h2>
    <p>You requested a new verification link for the %s waitlist.</p>
    <p>Please verify your email by clicking the button below:</p>
    <p style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;">Verify Email</a>
    </p>
    <p>Your current position: <strong>#%d</strong></p>
    <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">Thanks,<br>The %s Team</p>
</body>
</html>`,
		firstName, appName, data.VerificationURL, data.Position, appName)

	return Message{
		To:       []string{data.Email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}

// BuildOTPEmail creates an OTP verification email message.
// language: "en" for English or "fa" for Persian
func BuildOTPEmail(email string, code string, language string, expiryMinutes int) Message {
	const appName = "Simorgh"

	var subject, greeting, line1, line2, line3, codeLabel, expires, closing string

	if language == "fa" {
		subject = "کد تایید شما | Your Verification Code"
		greeting = "سلام,"
		line1 = "درخواست تایید هویت شما برای دسترسی به Simorgh دریافت شد."
		line2 = "لطفا برای تأیید هویت خود از کد زیر استفاده کنید:"
		line3 = "اگر این درخواست را نکردید، لطفا این ایمیل را무시 کنید."
		codeLabel = "کد تایید:"
		expires = fmt.Sprintf("این کد برای %d دقیقه معتبر است.", expiryMinutes)
		closing = "تیم Simorgh"
	} else {
		subject = "Your Verification Code | کد تایید شما"
		greeting = "Hi,"
		line1 = "You've requested to verify your identity for accessing Simorgh."
		line2 = "Please use the code below to verify your identity:"
		line3 = "If you didn't request this, please ignore this email."
		codeLabel = "Verification Code:"
		expires = fmt.Sprintf("This code is valid for %d minutes.", expiryMinutes)
		closing = "The Simorgh Team"
	}

	textBody := fmt.Sprintf(`%s

%s

%s

%s

%s

%s

%s

%s`, greeting, line1, line2, codeLabel, code, expires, line3, closing)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #2563eb;">%s</h2>
    <p>%s</p>
    <p>%s</p>
    <p style="text-align: center; margin: 30px 0; background-color: #f3f4f6; padding: 20px; border-radius: 6px;">
        <span style="font-size: 12px; color: #6b7280;">%s</span><br>
        <span style="font-size: 36px; font-weight: bold; font-family: monospace; color: #000; letter-spacing: 4px;">%s</span>
    </p>
    <p style="color: #ef4444; font-size: 14px; text-align: center;">%s</p>
    <p>%s</p>
    <p style="color: #6b7280; font-size: 14px; margin-top: 30px; border-top: 1px solid #e5e7eb; padding-top: 20px;">
        %s
    </p>
</body>
</html>`, greeting, line1, line2, codeLabel, code, expires, line3, closing)

	return Message{
		To:       []string{email},
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}
