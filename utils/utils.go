package utils

import "strings"

func ParseEmail(msg string) (subject, body string) {
	// Split the message into headers and body
	parts := strings.SplitN(msg, "\r\n\r\n", 2)
	headers := parts[0]
	if len(parts) > 1 {
		body = parts[1]
	}

	// Extract subject from headers
	for _, line := range strings.Split(headers, "\r\n") {
		if strings.HasPrefix(line, "Subject: ") {
			subject = strings.TrimPrefix(line, "Subject: ")
			break
		}
	}

	return
}
