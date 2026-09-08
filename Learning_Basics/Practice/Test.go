package main

import(
    "net/http"
    "encoding/json"
    "sync"
    "strconv"
    "strings"
)


type Entry struct {
    ID int
    Text string
    Mood string
    Category string
}


var (
    entries = make(map[int]*Entry)
    nextID =  1
    mu       sync.Mutex
)

func EntryHandler(w http.ResponseWriter, r *http.Request) {
    if(r.Method == http.MethodPost) {
        if r.URL.Path != "/entry/" {
            http.Error(w, "Invalid Path", http.StatusBadRequest)
            return
        }

        var entry Entry
        err := json.NewDecoder(r.Body).Decode(&entry)
        if err != nil {
            http.Error(w, "failed to decode the body", http.StatusBadRequest)
            return
        }

        mu.Lock()
        entry.ID = nextID
        entries[nextID] = &entry
        nextID++
        mu.Unlock()

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entry)
        return
    }

    if(r.Method == http.MethodGet) {
        
        urlString := r.URL.Path
        urlList := strings.Split(urlString, "/")

        if len(urlList) != 3 || urlList[2] == "" {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        id, err := strconv.Atoi(urlList[2])
        if err != nil {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        mu.Lock()
        entry, ok := entries[id]
        mu.Unlock()

        if !ok {
            http.Error(w, "Entry not found", http.StatusNotFound)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entry)
        return
    }
    if(r.Method == http.MethodPatch) {

        var entry Entry
        err := json.NewDecoder(r.Body).Decode(&entry)

        
		if err != nil {
			http.Error(w, "Failed to decode body", http.StatusBadRequest)
			return
		}

        urlString := r.URL.Path
        urlList := strings.Split(urlString, "/")

        if len(urlList) != 3 || urlList[2] == "" {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        id, err := strconv.Atoi(urlList[2])
        if err != nil {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        mu.Lock()        
        defer mu.Unlock()

        entryData, ok := entries[id]

        if !ok {
            http.Error(w, "Entry not found", http.StatusNotFound)
            return
        }

        entryData.Text = entry.Text
        entryData.Mood = entry.Mood
        entryData.Category = entry.Category

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entryData)
        return
    }
    if(r.Method == http.MethodDelete) {
        urlString := r.URL.Path
        urlList := strings.Split(urlString, "/")

        if len(urlList) != 3 || urlList[2] == "" {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        id, err := strconv.Atoi(urlList[2])
        if err != nil {
            http.Error(w, "Invalid ID", http.StatusBadRequest)
            return
        }

        mu.Lock()
        defer mu.Unlock()

        _, ok := entries[id]
        if !ok {
            http.Error(w, "Entry not found", http.StatusNotFound)
            return
        }

        delete(entries, id)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]bool {"deleted": true})
        return
    }
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func main() {
    
    http.HandleFunc("/entry/", EntryHandler)
    http.ListenAndServe(":8080", nil)    
}

