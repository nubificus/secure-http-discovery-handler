package verify

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Fetch a DER-encoded certificate from a URL and convert it to PEM
func fetchAndConvertToPEM(sourceURL string) (string, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body failed: %w", err)
	}

	// Find first byte 0x30 (start of ASN.1 SEQUENCE)
	idx := bytes.IndexByte(rawData, 0x30)
	if idx == -1 {
		return "", fmt.Errorf("could not locate DER start")
	}
	derData := rawData[idx:]

	cert, err := x509.ParseCertificate(derData)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	return strings.TrimSpace(string(pemBytes)), nil
}
func postPEM(pemData string, destURL string) error {
	req, err := http.NewRequest("POST", destURL, bytes.NewBufferString(pemData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to POST PEM data: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}
	return nil
}

func VerifyDevice(deviceIP string, diceIP string) (bool, error) {
	endpoint := fmt.Sprintf("http://%s/onboard", deviceIP)
	pemData, err := fetchAndConvertToPEM(endpoint)
	if err != nil {
		return false, err
	}
	err = postPEM(pemData, diceIP)
	if err != nil {
		return false, err
	}
	return true, nil
}
