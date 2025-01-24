package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func GenerateVideoToken(accountSid, apiKeySid, apiKeySecret, identity, roomName string) (string, error) {
	claims := jwt.MapClaims{
		"iss": apiKeySid,
		"sub": accountSid,
		"exp": time.Now().Add(6 * time.Hour).Unix(),
		"jti": fmt.Sprintf("%s-%d", apiKeySid, time.Now().UnixNano()),
		"grants": jwt.MapClaims{
			"identity": identity,
			"video": jwt.MapClaims{
				"room":      roomName,
				"room_type": "group",
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["cty"] = "twilio-fpa;v=1"
	return token.SignedString([]byte(apiKeySecret))
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		rand.New(rand.NewSource(time.Now().UnixNano()))
		identity := fmt.Sprintf("user-%04d", rand.Intn(10000))

		token, err := GenerateVideoToken(
			os.Getenv("TWILIO_ACCOUNT_SID"),
			os.Getenv("TWILIO_API_KEY_SID"),
			os.Getenv("TWILIO_API_KEY_SECRET"),
			identity,
			"group-video-room",
		)

		if err != nil {
			http.Error(w, `{"error": "Token generation failed"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(TokenResponse{Token: token})
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(router)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
