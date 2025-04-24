package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cafeList = map[string][]string{
	"moscow": {"Мир кофе", "Сладкоежка", "Кофе и завтраки", "Сытый студент"},
}

func mainHandle(w http.ResponseWriter, req *http.Request) {
	countStr := req.URL.Query().Get("count")
	if countStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("count missing"))
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong count value"))
		return
	}

	city := req.URL.Query().Get("city")

	cafe, ok := cafeList[city]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong city value"))
		return
	}

	if count > len(cafe) {
		count = len(cafe)
	}

	answer := strings.Join(cafe[:count], ",")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(answer))
}

// корректный запрос
func TestMainHandler_ValidRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cafe?count=2&city=moscow", nil)
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(respRec, req)

	require.Equal(t, http.StatusOK, respRec.Code)
	assert.NotEmpty(t, respRec.Body.String())
}

// город не поддерживается
func TestMainHandler_InvalidCity(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cafe?count=2&city=paris", nil)
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(respRec, req)

	require.Equal(t, http.StatusBadRequest, respRec.Code)
	assert.Equal(t, "wrong city value", respRec.Body.String())
}

// count больше, чем доступно
func TestMainHandlerWhenCountMoreThanTotal(t *testing.T) {
	totalCount := len(cafeList["moscow"])
	countMore := totalCount + 10

	params := url.Values{}
	params.Set("count", strconv.Itoa(countMore))
	params.Set("city", "moscow")

	req := httptest.NewRequest(http.MethodGet, "/cafe?"+params.Encode(), nil)
	respRec := httptest.NewRecorder()

	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(respRec, req)

	require.Equal(t, http.StatusOK, respRec.Code)

	cafes := strings.Split(respRec.Body.String(), ",")
	assert.Len(t, cafes, totalCount)
}
