package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DisputeLog matches the SQLite database transaction storage layout
type DisputeLog struct {
	ID           int       `json:"id"`
	SenderBank   string    `json:"sender_bank"`
	ReceiverBank string    `json:"receiver_bank"`
	Amount       float64   `json:"amount"`
	SessionID    string    `json:"session_id"`
	TransferDate time.Time `json:"transfer_date"`
	CanEscalate  bool      `json:"can_escalate"`
}

// Master CBN Banking & Digital Fintech Customer Care Email Directory
var customerCareDirectory = map[string]string{
	// Commercial Banks
	"access bank":           "cc@accessbankplc.com",
	"fidelity bank":         "true.serve@fidelitybank.ng",
	"first bank":            "firstcontact@firstbanknigeria.com",
	"fcmb":                  "customerservice@fcmb.com",
	"gtbank":                "help@gtbank.com",
	"heritage bank":         "customercare@hbng.com",
	"keystone bank":         "contactcentre@keystonebankng.com",
	"optimus bank":          "customercare@optimusbank.com",
	"premiumtrust bank":     "customercare@premiumtrustbank.com",
	"providus bank":         "customercare@providusbank.com",
	"polaris bank":          "yescenter@polarisbankng.com",
	"stanbic ibtc":          "customercarenigeria@stanbicibtc.com",
	"standard chartered":    "callcentre.ng@sc.com",
	"sterling bank":         "customercare@sterling.ng",
	"suntrust bank":         "customercare@suntrustbank.com",
	"union bank":            "customerservice@unionbankng.com",
	"uba":                   "cfc@ubagroup.com",
	"unity bank":            "customercare@unitybankng.com",
	"wema bank":             "purpleconnect@wemabank.com",
	"zenith bank":           "zenithdirect@zenithbank.com",
	
	// Major Digital Banks & Fintech Apps
	"opay":                  "customerservice@opay-inc.com",
	"kuda bank":             "help@kudabank.com",
	"moniepoint":            "support@moniepoint.com",
	"palmpay":               "support@palmpay.com",
	"vfd microfinance":      "support@vbank.ng",
	"fairmoney":             "help@fairmoney.ng",
	"carbon":                "customer-care@getcarbon.co",
	"rubies bank":           "hello@rubies.ng",
	"sparkle":               "support@sparkle.ng",
	"gomoney":               "support@gomoney.global",
	"piggyvest":             "contact@piggyvest.com",
	"brass":                 "help@brass.co",
	"pocketapp":             "hello@pocketapp.com",

	// Telecom Payment Service Banks (PSBs)
	"momo psb":              "customercare@momopsb.co.ng",
	"smartcash psb":         "customercare@://xlte.com",
	"9psb":                  "support@9psb.com.ng",
}

// LogTransactionHandler processes incoming JSON transaction logs with data validation
func LogTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "HTTP method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload DisputeLog
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Malformed JSON input payload structure", http.StatusBadRequest)
		return
	}

	// Clean trailing white spaces
	payload.SenderBank = strings.TrimSpace(payload.SenderBank)
	payload.ReceiverBank = strings.TrimSpace(payload.ReceiverBank)
	payload.SessionID = strings.TrimSpace(payload.SessionID)

	// 1. Financial Validation
	if payload.Amount <= 0 {
		http.Error(w, "Validation Error: Transaction amount must be greater than zero Naira (₦0).", http.StatusBadRequest)
		return
	}

	// 2. Empty Field Checks
	if payload.SenderBank == "" || payload.ReceiverBank == "" {
		http.Error(w, "Validation Error: Sender Bank and Receiver Bank are required fields.", http.StatusBadRequest)
		return
	}

	// 3. Strict 30-Digit NIBSS Reference Check
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if len(payload.SessionID) != 30 || !numericRegex.MatchString(payload.SessionID) {
		http.Error(w, "Validation Error: An authentic NIBSS Session ID must contain exactly 30 numeric digits.", http.StatusBadRequest)
		return
	}

	if payload.TransferDate.IsZero() {
		payload.TransferDate = time.Now()
	}

	// Write record directly into the database engine
	query := `INSERT INTO transaction_logs (sender_bank, receiver_bank, amount, session_id, transfer_date) VALUES (?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, payload.SenderBank, payload.ReceiverBank, payload.Amount, payload.SessionID, payload.TransferDate)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Validation Error: A dispute record with this specific Session ID has already been logged.", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Data storage error: %v", err), http.StatusInternalServerError)
		return
	}

	lastID, _ := result.LastInsertId()
	payload.ID = int(lastID)
	
	if time.Since(payload.TransferDate) > (48 * time.Hour) {
		payload.CanEscalate = true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payload)
}

// FetchAllTransactionsHandler queries and streams all logged transfers from SQLite
func FetchAllTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "HTTP method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := DB.Query("SELECT id, sender_bank, receiver_bank, amount, session_id, transfer_date FROM transaction_logs ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Failed to retrieve logs from database storage", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var collection []DisputeLog
	for rows.Next() {
		var item DisputeLog
		if err := rows.Scan(&item.ID, &item.SenderBank, &item.ReceiverBank, &item.Amount, &item.SessionID, &item.TransferDate); err != nil {
			continue
		}
		
		// Dynamically compute escalation availability relative to current real time
		if time.Since(item.TransferDate) > (48 * time.Hour) {
			item.CanEscalate = true
		}
		collection = append(collection, item)
	}

	if collection == nil {
		collection = []DisputeLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collection)
}

// CompileEscalationHandler compiles the data template into an email structure
func CompileEscalationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "HTTP method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid query ID parameter assignment", http.StatusBadRequest)
		return
	}

	var item DisputeLog
	query := "SELECT id, sender_bank, receiver_bank, amount, session_id, transfer_date FROM transaction_logs WHERE id = ?"
	err = DB.QueryRow(query, id).Scan(&item.ID, &item.SenderBank, &item.ReceiverBank, &item.Amount, &item.SessionID, &item.TransferDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Dispute entry missing from historical records database", http.StatusNotFound)
		return
	}

	if time.Since(item.TransferDate) <= (48 * time.Hour) {
		http.Error(w, "Regulatory Policy Error: Dispute falls within standard 48-hour resolution frame.", http.StatusForbidden)
		return
	}

	bankKey := strings.ToLower(item.SenderBank)
	targetEmail, exists := customerCareDirectory[bankKey]
	if !exists {
		targetEmail = "complaints@cbn.gov.ng" // Fallback straight to CBN Consumer Protection
	}

	emailSubject := fmt.Sprintf("URGENT: Unresolved Transaction Dispute / NIBSS Session ID: %s", item.SessionID)
	emailBody := fmt.Sprintf(
		"Dear Customer Operations Team,\n\n"+
			"I am filing a formal claim for an unreversed transaction debit that has broken the Central Bank of Nigeria's mandatory 48-hour timeline requirement.\n\n"+
			"Audit Reference Ledger:\n"+
			"- Source Financial Institution: %s\n"+
			"- Destination Institution: %s\n"+
			"- Value Amount: ₦%.2f\n"+
			"- Originating Timestamp: %s\n"+
			"- NIBSS Session ID Key: %s\n\n"+
			"Please run an immediate log check and process a complete fallback reversal. If this is not resolved, this data sheet will be filed directly on the CBN Consumer Protection portal.\n\n"+
			"Sincerely,\nClaimant User",
		item.SenderBank, item.ReceiverBank, item.Amount, item.TransferDate.Format("2006-01-02 15:04:05"), item.SessionID,
	)

	// Print compiled copy out to the workspace console stream logs
	fmt.Printf("\n=== DISPUTE EMAIL SENT TO: %s ===\nSubject: %s\n\n%s\n=================================\n", targetEmail, emailSubject, emailBody)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":          "compiled",
		"dispatched_to":   targetEmail,
		"complaint_email": emailBody,
	})
}
