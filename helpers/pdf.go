package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"text/template"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	qrcode "github.com/skip2/go-qrcode"
)

type RequestPdf struct {
	bodies []string
}

func (r *RequestPdf) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.bodies = append(r.bodies, buf.String())
	return nil
}

func (r *RequestPdf) GeneratePDF() (*bytes.Buffer, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}

	for _, body := range r.bodies {
		pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(body))))
	}

	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Buffer(), nil
}

func GenerateTicketsPDF(order *models.Order) (*bytes.Buffer, error) {
	r := RequestPdf{}

	for _, ticket := range order.Tickets {
		img, err := qrcode.New(fmt.Sprintf("%d-%d", order.ID, ticket.ID), qrcode.Medium)
		if err != nil {
			return nil, err
		}

		base64, err := EncodeImage(img.Image(256))
		if err != nil {
			return nil, err
		}

		fmt.Println(string(base64))

		if err := r.ParseTemplate("./templates/ticket.html", models.TicketHTML{
			ID:                 ticket.ID,
			Firstname:          RemoveAccents(order.Client.Firstname),
			Lastname:           order.Client.Lastname,
			EventStartDateTime: ticket.Event.StartDateTime.String(),
			EventEndDateTime:   ticket.Event.EndDateTime.String(),
			Price:              ticket.Event.Price,
			Image:              base64,
		}); err != nil {
			return nil, err
		}
	}

	mem, err := r.GeneratePDF()
	if err != nil {
		return nil, err
	}

	return mem, nil
}

func EncodeImage(m image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
