package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/models"
	"github.com/dev-bruno-arruda/api-pedidos/api_service/pgk/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Product == "" {
		http.Error(w, "Campo 'product' é obrigatório", http.StatusBadRequest)
		return
	}

	if req.Quantity < 1 {
		http.Error(w, "Campo 'quantity' deve ser maior que 0", http.StatusBadRequest)
		return
	}

	response, err := h.service.CreateOrder(r.Context(), req)
	if err != nil {
		log.Printf("Erro ao criar pedido: %v", err)
		http.Error(w, "Erro ao criar pedido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("Pedido criado com sucesso: %s", response.OrderID)
}
