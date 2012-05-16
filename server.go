package main

import (
	"encoding/json"
	"net/http"
	"sync"
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
		s, ok := c.Stats[k]
		if !ok {
			http.Error(w, "unknown stat: "+k, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err := enc.Encode(s.Values)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
	}
	wg.Wait()
}
