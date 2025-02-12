package buy

import (
	"fmt"
	"net/http"
)

type Buyer interface {
	Buy(name string) (string, error)
}

type Request struct {
	Name string `json:"name,omitempty"`
}

type Response struct {
	Name string `json:"name,omitempty"`
}

func Buy(w http.ResponseWriter, r *http.Request, b Buyer) {
	item := r.URL.Path[len("/api/buy")+1:]
	price, err := b.Buy(item)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	fmt.Fprintf(w, "Name: %s, Price: %s", item, price)
}
