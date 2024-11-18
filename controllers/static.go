package controllers

import (
	"net/http"
)

func StaticHandler(t template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, r, nil)
	}
}

func FAQ(t template) http.HandlerFunc {
	questions := []struct {
		Q string
		A string
	}{
		{
			Q: "Do you have Privacy Agreement?",
			A: "No, we don't. We collect all data and sell it.",
		},
		{
			Q: "How can I contact you?",
			A: "We dont't provide any. Trust in us.",
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, r, questions)
	}
}
