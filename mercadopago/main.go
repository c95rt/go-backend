package mercadopago

import (
	"bytes"
	"encoding/json"
	"fmt"
	io "io/ioutil"
	"net/http"
	"strconv"

	"bitbucket.org/parqueoasis/backend/models"
	shortuuid "github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"
)

const (
	mpContentType = `application/json`
)

type MP struct {
	BaseURL          string
	Token            string
	NotificationPath string
	PathPreferences  string
	GetPaymentURL    string
}

var MPActions interface {
	MPCreatePreference()
}

type MPCreatePreferenceRequest struct {
	NotificationURL   string               `json:"notification_url"`
	ExternalReference string               `json:"external_reference"`
	Items             []MPPreferenceItem   `json:"items"`
	BackUrls          MPPreferenceBackUrls `json:"back_urls"`
}

type MPPreferenceBackUrls struct {
	Success string `json:"success"`
	Failure string `json:"faillure"`
}

type MPPreferenceItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int    `json:"unit_price"`
}

type MPCreatePreferenceResponse struct {
	InitPoint         string `json:"init_point"`
	ExternalReference string `json:"external_reference"`
}

type MPGetPaymentReponse struct {
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference"`
}

func (mp *MP) MPCreatePreference(order *models.Order, baseURL string) (*MPCreatePreferenceResponse, error) {
	requestBody := MPCreatePreferenceRequest{
		NotificationURL:   fmt.Sprintf("%s%s", baseURL, mp.NotificationPath),
		ExternalReference: shortuuid.New(),
		BackUrls: MPPreferenceBackUrls{
			Success: "https://dev.parqueoasis.cl/checkout/success",
			Failure: "https://dev.parqueoasis.cl/checkout/failure",
		},
	}

	item := MPPreferenceItem{
		ID:          strconv.Itoa(order.ID),
		Title:       "Entrada Parque",
		Description: fmt.Sprintf("%s-%s", order.Event.StartDateTime.String(), order.Event.EndDateTime.String()),
		Quantity:    order.Tickets,
		UnitPrice:   order.Event.Price,
	}

	requestBody.Items = append(requestBody.Items, item)

	responseBody, err := mpPost(fmt.Sprintf("%s%s?access_token=%s", mp.BaseURL, mp.PathPreferences, mp.Token), &requestBody)
	if err != nil {
		return nil, err
	}

	if responseBody == nil {
		return nil, errors.New("failed creating preference in Mercado Pago")
	}

	var response MPCreatePreferenceResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (mp *MP) MPGetPayment(id string) (*MPGetPaymentReponse, error) {
	responseBody, err := mpGet(fmt.Sprintf("%s%s?access_token=%s", mp.GetPaymentURL, id, mp.Token))
	if err != nil {
		return nil, err
	}

	if responseBody == nil {
		return nil, errors.New("failed creating preference in Mercado Pago")
	}

	var response MPGetPaymentReponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func mpPost(url string, body interface{}) ([]byte, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	response, err := http.Post(url, mpContentType, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("bad response %d", response.StatusCode)
	}

	return responseBody, nil
}

func mpGet(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("bad response %d", response.StatusCode)
	}

	return responseBody, nil
}
