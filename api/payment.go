package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/mux"
	"github.com/lithammer/shortuuid/v3"
	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/govalidator"
)

var insertPaymentRules = govalidator.MapData{
	"method_id": []string{"required", "numeric"},
}

func InsertPaymentMercadoPago(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsClient {
		w.WriteJSON(http.StatusInternalServerError, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["order_id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed parsing order id")
		return
	}

	order, err := ctx.DB.GetOrderByID(orderID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting order")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "order not found")
		return
	}

	if order.Client.ID != userInfo.ID {
		w.WriteJSON(http.StatusForbidden, nil, err, "invalid user")
		return
	}

	if order.Payment != nil {
		if order.Payment.Status != nil {
			if order.Payment.Status.ID == db.ConstPaymentStatuses.Approved.ID {
				w.WriteJSON(http.StatusBadRequest, nil, err, "order already paid")
				return
			}
			if order.Payment.Status.ID == db.ConstPaymentStatuses.Processing.ID {
				w.WriteJSON(http.StatusBadRequest, nil, err, "already processing payment")
				return
			}
		}
	}

	for _, ticket := range order.Tickets {
		order.Price += ticket.Event.Price
	}

	response, err := ctx.MercadoPago.MPCreatePreference(order, ctx.Config.BackendBaseURL)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "problems with Mercado Pago")
		return
	}

	if response == nil {
		w.WriteJSON(http.StatusInternalServerError, nil, nil, "failed creating payment in Mercado Pago")
		return
	}

	if response.ExternalReference == "" {
		w.WriteJSON(http.StatusInternalServerError, nil, nil, "bad response from Mercado Pago")
		return
	}

	newOpts := db.InsertPaymentOpts{
		MethodID:     db.ConstPaymentMethods.MercadoPago.ID,
		Amount:       order.Price,
		UserID:       userInfo.ID,
		OrderID:      order.ID,
		PreferenceID: response.ExternalReference,
		StatusID:     db.ConstPaymentStatuses.Created.ID,
	}

	_, err = ctx.DB.InsertPayment(&newOpts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed inserting payment")
		return
	}

	w.WriteJSON(http.StatusOK, response, nil, "")
	return
}

type UpdatePaymentMercadoPagoOpts struct {
	Data *UpdatePaymentMercadoPagoDataOpts `json:"data"`
}

type UpdatePaymentMercadoPagoDataOpts struct {
	ID string `json:"id"`
}

func UpdatePaymentMercadoPago(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.StartLogger("UpdatePaymentMercadoPago")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.LogError(err, "failed reading body")
		return
	}
	defer r.Body.Close()

	var opts UpdatePaymentMercadoPagoOpts
	if err := json.Unmarshal(body, &opts); err != nil {
		w.LogError(err, "failed unmarshaling body")
		return
	}

	if opts.Data == nil {
		w.LogError(nil, "no data")
		return
	}

	response, err := ctx.MercadoPago.MPGetPayment(opts.Data.ID)
	if err != nil {
		w.LogError(err, "failed getting payment from Mercado Pago")
		return
	}

	paymentStatus, err := ctx.DB.GetPaymentStatusByMethodIDAndMethodStatusName(db.ConstPaymentMethods.MercadoPago.ID, response.Status)
	if err != nil {
		w.LogError(err, "failed getting status")
		return
	}

	if paymentStatus == nil {
		w.LogError(nil, "payment status not found")
		return
	}

	if err := ctx.DB.UpdatePaymentStatus(response.ExternalReference, paymentStatus.ID); err != nil {
		w.LogError(err, "failed updating payment")
		return
	}

	go func(ctx *config.AppContext, externalReference string) {
		order, err := ctx.DB.GetOrderByExternalReference(externalReference)
		if err != nil {
			w.LogError(err, "failed getting order")
			return
		}

		for _, ticket := range order.Tickets {
			order.Price += ticket.Event.Price
		}

		pdfBuffer, err := helpers.GenerateTicketsPDF(order)
		if err != nil {
			w.LogError(err, "failed generating PDF")
			return
		}

		ed := &helpers.EmailData{
			EmailTo:      order.Client.Email,
			NameTo:       order.Client.Firstname,
			EmailFrom:    ctx.Config.Mail.EmailFrom,
			NameFrom:     ctx.Config.Mail.NameFrom,
			Subject:      ctx.Config.Mail.PaymentSuccess.Subject,
			TemplatePath: fmt.Sprintf("%s%s/%s", ctx.Config.Mail.Folder, ctx.Config.Mail.Path, ctx.Config.Mail.PaymentSuccess.Template),
			FileName:     ctx.Config.Mail.PaymentSuccess.FileName,
			FileContent:  pdfBuffer.Bytes(),
			AwsSMTP:      ctx.AwsSMTP,
		}

		err = ed.SendEmail(models.OrderHTML{
			ID:            order.ID,
			Firstname:     order.Client.Firstname,
			Lastname:      order.Client.Lastname,
			PaymentMethod: db.ConstPaymentMethods.MercadoPago.Name,
			OrderPrice:    order.Price,
		})
		if err != nil {
			w.LogError(err, "failed sending email")
			return
		}

		w.LogInfo(nil, "success sending email")
	}(ctx, response.ExternalReference)

	w.LogInfo(response, "success")
	return
}

func InsertPaymentCashier(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["order_id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed parsing order id")
		return
	}

	order, err := ctx.DB.GetOrderByID(orderID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting order")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "order not found")
		return
	}

	if order.Payment != nil {
		if order.Payment.Status != nil {
			if order.Payment.Status.ID == db.ConstPaymentStatuses.Approved.ID {
				w.WriteJSON(http.StatusBadRequest, nil, err, "order already paid")
				return
			}
			if order.Payment.Status.ID == db.ConstPaymentStatuses.Processing.ID {
				w.WriteJSON(http.StatusBadRequest, nil, err, "already processing payment")
				return
			}
		}
	}

	for _, ticket := range order.Tickets {
		order.Price += ticket.Event.Price
	}

	newOpts := db.InsertPaymentOpts{
		MethodID:     db.ConstPaymentMethods.MercadoPago.ID,
		Amount:       order.Price,
		UserID:       userInfo.ID,
		OrderID:      order.ID,
		PreferenceID: shortuuid.New(),
		StatusID:     db.ConstPaymentStatuses.Approved.ID,
	}

	_, err = ctx.DB.InsertPayment(&newOpts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed inserting payment")
		return
	}

	pdfBuffer, err := helpers.GenerateTicketsPDF(order)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed generating pdfs")
		return
	}

	url, err := helpers.AddFileToS3(ctx, pdfBuffer, fmt.Sprintf("%s/%d.pdf", ctx.Config.AwsS3.S3PathTicket, order.ID))
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed uploading pdf")
		return
	}

	go func(ctx *config.AppContext, order *models.Order, pdfBuffer *bytes.Buffer) {
		ed := &helpers.EmailData{
			EmailTo:      order.Client.Email,
			NameTo:       order.Client.Firstname,
			EmailFrom:    ctx.Config.Mail.EmailFrom,
			NameFrom:     ctx.Config.Mail.NameFrom,
			Subject:      ctx.Config.Mail.PaymentSuccess.Subject,
			TemplatePath: fmt.Sprintf("%s%s/%s", ctx.Config.Mail.Folder, ctx.Config.Mail.Path, ctx.Config.Mail.PaymentSuccess.Template),
			FileName:     ctx.Config.Mail.PaymentSuccess.FileName,
			FileContent:  pdfBuffer.Bytes(),
			AwsSMTP:      ctx.AwsSMTP,
		}

		err = ed.SendEmail(models.OrderHTML{
			ID:            order.ID,
			Firstname:     order.Client.Firstname,
			Lastname:      order.Client.Lastname,
			PaymentMethod: db.ConstPaymentMethods.Cashier.Name,
			OrderPrice:    order.Price,
		})
		if err != nil {
			w.LogError(err, "failed sending email")
			return
		}
		w.LogInfo(nil, "success sending email")
	}(ctx, order, pdfBuffer)

	w.WriteJSON(http.StatusOK, models.TicketPDF{
		URL: url,
	}, nil, "")
	return
}
