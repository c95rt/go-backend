package helpers

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"log"
	"strings"
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

const (
	ConstHTMLNewPage = `
	<div class="new-page"></div>
	`
)

func (r *RequestPdf) GeneratePDF() (*bytes.Buffer, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}

	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(strings.Join(r.bodies, ConstHTMLNewPage)))))

	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Buffer(), nil
}

func GenerateOrderPDF(order *models.Order) (*bytes.Buffer, error) {
	r := RequestPdf{}

	img, err := qrcode.New(order.TransactionID, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	base64, err := EncodeImage(img.Image(256))
	if err != nil {
		return nil, err
	}

	if err := r.ParseTemplate("./templates/pdf/order.html", models.OrderPDFHTML{
		ID:            order.ID,
		Firstname:     RemoveAccents(order.Client.Firstname),
		Lastname:      order.Client.Lastname,
		Date:          order.Event.StartDateTime.Format("02-01-2006"),
		EventType:     order.Event.Type.Name,
		Price:         order.Price,
		Image:         base64,
		TransactionID: order.TransactionID,
		Tickets:       order.Tickets,
	}); err != nil {
		return nil, err
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
