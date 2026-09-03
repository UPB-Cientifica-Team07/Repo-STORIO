package streaming

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// =====================================
// CLIENT
// =====================================

type Client struct {
	baseURL    string
	httpClient *http.Client
}

// =====================================
// RESPUESTA
// =====================================

type RegisterVideoResult struct {
	Success         bool
	Message         string
	VideoID         string
	FileID          string
	DurationSeconds int
	Quality         string
	Format          string
}

// =====================================
// SOAP RESPONSE
// =====================================

type registerVideoEnvelope struct {
	Body registerVideoBody `xml:"Body"`
}

type registerVideoBody struct {
	Response *registerVideoResponse `xml:"registerVideoResponse"`
	Fault    *soapFault             `xml:"Fault"`
}

type registerVideoResponse struct {
	Success         bool   `xml:"success"`
	Message         string `xml:"message"`
	VideoID         string `xml:"idVideo"`
	FileID          string `xml:"idArchivo"`
	DurationSeconds int    `xml:"duracionSegundos"`
	Quality         string `xml:"calidad"`
	Format          string `xml:"formato"`
}

type soapFault struct {
	Code    string `xml:"faultcode"`
	Message string `xml:"faultstring"`
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewClient(
	baseURL string,
) *Client {

	return &Client{
		baseURL: strings.TrimRight(
			baseURL,
			"/",
		),

		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// =====================================
// REGISTER VIDEO
// =====================================

func (c *Client) RegisterVideo(
	fileID string,
) (*RegisterVideoResult, error) {

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {
		return nil,
			fmt.Errorf(
				"el file ID es obligatorio",
			)
	}

	payload :=
		fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope
    xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:tns="urn:UPBCientificaStreaming">
    <soap:Body>
        <tns:registerVideoRequest>
            <tns:idArchivo>%s</tns:idArchivo>
        </tns:registerVideoRequest>
    </soap:Body>
</soap:Envelope>`,
			xmlEscape(
				fileID,
			),
		)

	request,
		err :=
		http.NewRequest(
			http.MethodPost,
			c.baseURL+"/",
			bytes.NewBufferString(
				payload,
			),
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudo construir petición SOAP: %w",
				err,
			)
	}

	request.Header.Set(
		"Content-Type",
		"text/xml; charset=utf-8",
	)

	request.Header.Set(
		"SOAPAction",
		`"urn:UPBCientificaStreaming#registerVideo"`,
	)

	response,
		err :=
		c.httpClient.Do(
			request,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error conectando con Streaming Service: %w",
				err,
			)
	}

	defer response.Body.Close()

	body,
		err :=
		io.ReadAll(
			response.Body,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error leyendo respuesta de Streaming Service: %w",
				err,
			)
	}

	var envelope registerVideoEnvelope

	if err :=
		xml.Unmarshal(
			body,
			&envelope,
		); err != nil {

		return nil,
			fmt.Errorf(
				"respuesta SOAP inválida de Streaming Service: %w",
				err,
			)
	}

	if envelope.Body.Fault != nil {

		return nil,
			fmt.Errorf(
				"Streaming SOAP Fault %s: %s",
				envelope.Body.Fault.Code,
				envelope.Body.Fault.Message,
			)
	}

	if response.StatusCode !=
		http.StatusOK {

		return nil,
			fmt.Errorf(
				"Streaming Service respondió HTTP %d",
				response.StatusCode,
			)
	}

	if envelope.Body.Response == nil {

		return nil,
			fmt.Errorf(
				"Streaming Service no devolvió registerVideoResponse",
			)
	}

	result :=
		envelope.Body.Response

	if !result.Success {

		return nil,
			fmt.Errorf(
				"Streaming rechazó registro: %s",
				result.Message,
			)
	}

	return &RegisterVideoResult{
		Success:         result.Success,
		Message:         result.Message,
		VideoID:         result.VideoID,
		FileID:          result.FileID,
		DurationSeconds: result.DurationSeconds,
		Quality:         result.Quality,
		Format:          result.Format,
	}, nil
}

// =====================================
// XML ESCAPE
// =====================================

func xmlEscape(
	value string,
) string {

	var buffer bytes.Buffer

	_ =
		xml.EscapeText(
			&buffer,
			[]byte(value),
		)

	return buffer.String()
}
