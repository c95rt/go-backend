package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/govalidator"
)

func InsertOrder(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	// timeLocation, err := time.LoadLocation("America/Santiago")
	// if err != nil {
	// 	w.WriteJSON(http.StatusInternalServerError, nil, err, "failed loading time location")
	// }

	if !userInfo.IsAdmin && !userInfo.IsCashier && !userInfo.IsClient {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	var opts models.InsertOrdersOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertOrdersRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	userID := userInfo.ID
	if userInfo.IsClient {
		userID = 1
	}

	event, err := ctx.DB.GetEventByID(opts.EventID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if event == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "Evento no encontrado")
		return
	}

	if event.EndDateTime.Before(time.Now()) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "El evento ya ha terminado")
		return
	}

	order, err := ctx.DB.InsertOrder(userID, opts.UserID, event.ID, opts.Tickets, event.Price)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	order.Event = event
	order.Price = event.Price * opts.Tickets

	w.WriteJSON(http.StatusOK, order, nil, "")
}

func GetOrders(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier && !userInfo.IsClient {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol inválido")
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetOrdersRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var opts models.GetOrdersOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	orders, err := ctx.DB.GetOrders(&opts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	w.WriteJSON(http.StatusOK, orders, nil, "")
}

func GetOrderPDF(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier && !userInfo.IsClient {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol Inválido")
		return
	}

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	order, err := ctx.DB.GetOrderByID(orderID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "Orden no encontrada")
		return
	}

	if order != nil {
		if *order.Used == true {
			w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden ya está caducada")
			return
		}
	}

	if userInfo.IsClient {
		if order.Client.ID != userInfo.ID {
			w.WriteJSON(http.StatusForbidden, nil, nil, "El cliente no corresponde a la orden")
			return
		}
	}

	pdfBuffer, err := helpers.GenerateOrderPDF(order)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	url, err := helpers.AddFileToS3(ctx, pdfBuffer, fmt.Sprintf("%s/%d/%d.pdf", ctx.Config.AwsS3.S3PathOrder, userInfo.ID, order.ID))
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	w.WriteJSON(http.StatusOK, models.OrderPDF{
		URL: url,
	}, nil, "")
	return
}

func UseOrder(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol inválido")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	order, err := ctx.DB.GetOrderByID(id)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "Orden no encontrada")
		return
	}

	if order.Payment == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden no ha sido pagada")
		return
	}

	if order.Payment.Status == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden no ha sido pagada")
		return
	}

	if order.Payment.Status.ID != db.ConstPaymentStatuses.Approved.ID {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden no ha sido pagada")
		return
	}

	if order.Event == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden no tiene evento")
		return
	}

	if !time.Now().After(order.Event.StartDateTime) && !time.Now().Equal(order.Event.StartDateTime) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "El evento no ha empezado")
		return
	}

	if !time.Now().Before(order.Event.EndDateTime) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "El evento ya ha terminado")
		return
	}

	if order != nil {
		if *order.Used == true {
			w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden ya está caducada")
			return
		}
	}

	if err := ctx.DB.UseOrder(order.ID, userInfo.ID); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func UpdateOrder(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	var opts models.UpdateOrderOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.UpdateOrderRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	order, err := ctx.DB.GetOrderByID(orderID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "Orden no encontrada")
		return
	}

	if order != nil {
		if *order.Used == true {
			w.WriteJSON(http.StatusBadRequest, nil, nil, "La orden ya está caducada")
			return
		}
	}

	if err := ctx.DB.UpdateOrder(orderID, opts.EventID); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error actualizando el evento de la orden")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func GetSalesSummary(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol inválido")
		return
	}

	dailySales, err := ctx.DB.GetSalesSummary()
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if len(dailySales) == 0 {
		w.WriteJSON(http.StatusNotFound, nil, nil, "Todavía no hay ventas")
		return
	}

	var salesSummary models.SalesSummary
	monthlyCurrentYearSalesMap := make(map[string]int64)
	monthlyLastYearSalesMap := make(map[string]int64)
	currentYear, currentMonth, currentDay := time.Now().Date()

	for _, dailySale := range dailySales {
		dailySaleYear, dailySaleMonth, dailySaleDay := dailySale.Date.Date()
		if dailySaleYear == currentYear && dailySaleMonth == currentMonth && dailySaleDay == currentDay {
			salesSummary.CurrentDay += dailySale.Total
		}
		if dailySaleYear == currentYear && dailySaleMonth == currentMonth {
			salesSummary.CurrentMonth += dailySale.Total
		}
		if dailySaleYear == currentYear {
			salesSummary.CurrentYear += dailySale.Total
			monthlyCurrentYearSalesMap[dailySaleMonth.String()] += dailySale.Total
		}
		if dailySaleYear+1 == currentYear {
			monthlyLastYearSalesMap[dailySaleMonth.String()] += dailySale.Total
		}
	}

	for month, total := range monthlyCurrentYearSalesMap {
		salesSummary.MonthlyCurrentYear = append(salesSummary.MonthlyCurrentYear, models.MonthlySalesSummaryDetail{
			Year:  currentYear,
			Month: month,
			Total: total,
		})
	}

	for month, total := range monthlyLastYearSalesMap {
		salesSummary.MonthlyLastYear = append(salesSummary.MonthlyLastYear, models.MonthlySalesSummaryDetail{
			Year:  currentYear - 1,
			Month: month,
			Total: total,
		})
	}

	w.WriteJSON(http.StatusOK, salesSummary, nil, "")
}

func GetCashierSummary(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol inválido")
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetCashierSummaryRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validation")
		return
	}

	var opts models.GetCashierSummaryOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	if !userInfo.IsAdmin {
		opts.CashierID = userInfo.ID
	}

	monthlySales, err := ctx.DB.GetCashierSummary(opts.CashierID, opts.DateFrom, opts.DateTo)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if len(monthlySales) == 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "El cajero aún no realiza acciones")
		return
	}

	var summary models.CashierSummary

	monthlySalesMap := make(map[int]map[string]int64)
	monthlyUsesMap := make(map[int]map[string]int64)

	for _, monthlySale := range monthlySales {
		monthlySaleYear, monthlySaleMonth, _ := monthlySale.Date.Date()

		if _, ok := monthlySalesMap[monthlySaleYear]; !ok {
			monthlySalesMap[monthlySaleYear] = make(map[string]int64)
		}
		monthlySalesMap[monthlySaleYear][monthlySaleMonth.String()] += monthlySale.TotalSales

		if _, ok := monthlyUsesMap[monthlySaleYear]; !ok {
			monthlyUsesMap[monthlySaleYear] = make(map[string]int64)
		}
		monthlyUsesMap[monthlySaleYear][monthlySaleMonth.String()] += monthlySale.TotalUses
		summary.TotalSales += monthlySale.TotalSales
		summary.TotalUses += monthlySale.TotalUses
	}

	for year, monthlySalesByYearMap := range monthlySalesMap {
		for month, sales := range monthlySalesByYearMap {
			summary.MonthlySales = append(summary.MonthlySales, models.MonthlySalesSummaryDetail{
				Year:  year,
				Month: month,
				Total: sales,
			})
		}
	}

	for year, monthlyUsesByYearMap := range monthlyUsesMap {
		for month, uses := range monthlyUsesByYearMap {
			summary.MonthlyUses = append(summary.MonthlyUses, models.MonthlySalesSummaryDetail{
				Year:  year,
				Month: month,
				Total: uses,
			})
		}
	}

	w.WriteJSON(http.StatusOK, summary, nil, "")
}
