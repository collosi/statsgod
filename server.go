package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var (
	AccessControlAllowOrigin = "*"
)

func DataHandler(w http.ResponseWriter, r *http.Request, coreChan chan func(c *Core)) {
	k := r.FormValue("k")
	if k == "" {
		http.Error(w, "'k' required", http.StatusBadRequest)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	coreChan <- func(c *Core) {
		defer wg.Done()
		h := w.Header()
		if AccessControlAllowOrigin != "" {
			h.Set("Access-Control-Allow-Origin", AccessControlAllowOrigin)
		}
		h.Set("Content-Type", "application/json")
		s, ok := c.Stats[k]
		if !ok {
			http.Error(w, "{}", http.StatusNotFound)
			return
		}
		var values []Datum
		fmt.Printf("%v\n", s.Values)
		s.CopyValues(&values)
		fmt.Printf("%v\n", values)
		enc := json.NewEncoder(w)
		err := enc.Encode(values)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
	}
	wg.Wait()
}
