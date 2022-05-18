package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/jjlock/bank/currency"
	"github.com/jjlock/bank/db"
	"github.com/jjlock/bank/validate"
	"golang.org/x/crypto/bcrypt"
)

type server struct {
	db      db.Store
	session sessionStore
}

func New() *server {
	return &server{
		db:      make(db.Memcache),
		session: newCookieStore(),
	}
}

func (s *server) LoadDB(file string) error {
	return s.db.Load(file)
}

func (s *server) SaveDB(file string) error {
	return s.db.Save(file)
}

func (s *server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "/login")

	if r.Method == http.MethodGet {
		if _, err := s.session.getUser(r); err == nil {
			log.Println("User already logged in. Redirecting to /account")
			http.Redirect(w, r, "/account", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./templates/index.html")
	} else if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// get user account
		account, err := s.db.GetAccount(username)
		if err != nil {
			log.Println("Failed login:", err)
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}

		// compare password hashes
		if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
			log.Println(account.Username, "failed login:", err)
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}

		// create the session
		err = s.session.create(w, r, account.Username)
		if err != nil {
			log.Println("Failed to create session:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		log.Println(user(username), "logged in")
	} else {
		log.Println("Unsupported HTTP method received:" + r.Method + " /login")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "/signup")

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./templates/signup.html")
	} else if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		balanceString := r.FormValue("balance")

		// validate username and password
		if err := validate.Username(username); err != nil {
			log.Println("Invalid username:", err)
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}
		if err := validate.Password(password); err != nil {
			log.Println("Invalid password:", err)
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}

		// validate balance string
		balance, err := currency.ParseMicroUSD(balanceString)
		if err != nil {
			log.Println("Failed signup:", err)
			http.Error(w, "Invalid info", http.StatusBadRequest)
			return
		}

		// create the session
		err = s.session.create(w, r, username)
		if err != nil {
			log.Println("Failed to create session:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// add user account to database
		err = s.db.AddAccount(username, password, balance)
		if err != nil {
			log.Println("Failed signup:", err)
			http.Error(w, "Invalid info", http.StatusBadRequest)
			return
		}

		log.Println(user(username), "account created")
	} else {
		log.Println("Unsupported HTTP method received: " + r.Method + " /signup")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *server) AccountHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "/account")

	if r.Method != http.MethodGet {
		log.Println("Unsupported HTTP method received: " + r.Method + " /account")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// get user from session
	username, err := s.session.getUser(r)
	if err != nil {
		log.Println("Failed to get user from session:", err)
		here := url.QueryEscape("http://" + r.Host + r.URL.Path)
		http.Redirect(w, r, "/login?return_to="+here, http.StatusFound)
		return
	}

	// get user account
	account, err := s.db.GetAccount(username)
	if err != nil {
		log.Println("Failed to get account:", err)
		http.Error(w, "Invalid info", http.StatusBadRequest)
		return
	}

	// parse template files
	template, err := template.ParseFiles(filepath.Join("templates", "account.html"))
	if err != nil {
		log.Println("Failed to parse template:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// data to display on account page
	data := struct {
		Username string
		Balance  string
	}{
		account.Username,
		account.Balance.String(),
	}

	log.Println("Sending account info for", user(account.Username))
	template.Execute(w, data)
}

func (s *server) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "/transaction")

	if r.Method != http.MethodPost {
		log.Println("Unsupported HTTP method received: " + r.Method + " /transaction")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	amountString := r.FormValue("amount")
	transactionType := r.FormValue("type")

	// get username from session
	username, err := s.session.getUser(r)
	if err != nil {
		log.Println("Failed to get user from session:", err)
		http.Error(w, "Invalid info", http.StatusBadRequest)
		return
	}

	// convert amount string to micro USD
	amount, err := currency.ParseMicroUSD(amountString)
	if err != nil {
		log.Println("Could not parse amount:", err)
		http.Error(w, "Invalid info", http.StatusBadRequest)
		return
	}

	// determine whether the user is depositing or withdrawing money
	if transactionType == "deposit" {
		log.Println(user(username), "depositing:", amount)
		err = s.db.Deposit(username, amount)
		if err != nil {
			log.Println(user(username), "failed deposit:", err)
			http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
			return
		}
	} else if transactionType == "withdraw" {
		log.Println(user(username), "withdrawing:", amount)
		err = s.db.Withdraw(username, amount)
		if err != nil {
			log.Println(user(username), "failed withdrawl:", err)
			http.Error(w, "Invalid withdrawal amount", http.StatusBadRequest)
			return
		}
	} else {
		log.Println("Unknown transaction type")
		http.Error(w, "Unknown transaction type", http.StatusBadRequest)
		return
	}

	// get account to respond with updated balance
	account, err := s.db.GetAccount(username)
	if err != nil {
		log.Println("Could not retrieve account:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Println(user(account.Username), "updated balance:", account.Balance.String())
	w.Write([]byte(account.Balance.String()))
}

func (s *server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "/logout")

	if r.Method != http.MethodPost {
		log.Println("Unsupported HTTP method received: " + r.Method + " /logout")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// get user from session
	username, err := s.session.getUser(r)
	if err != nil {
		log.Println("Failed to get user from session:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// expire the session
	err = s.session.expire(w, r)
	if err != nil {
		log.Println("Failed to expire session:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Println(user(username), "logged out")
}

func user(username string) string {
	return fmt.Sprintf("[User: %q]", username)
}
