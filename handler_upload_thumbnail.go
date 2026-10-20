package main

import (
	"fmt"
	"net/http"
	"io"
	"mime"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}
	mediaTypes, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if mediaTypes != "image/jpeg" && mediaTypes != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", err)
		return
	}
	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get video", err)
		return
	}
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "IDs do not match", err)
		return
	}
	parts := strings.Split(mediaTypes, "/")
	someBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not read bytes", err)
	}
	fileExt := base64.RawURLEncoding.EncodeToString(someBytes)
	fullPath := filepath.Join(cfg.assetsRoot, fileExt) 
	newFile, err := os.Create(fullPath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not create file path", err)
		return
	}
	_, err = io.Copy(newFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not copy file", err)
		return
	}
	newURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileExt)
	videoMetadata.ThumbnailURL = &newURL
	err = cfg.db.UpdateVideo(videoMetadata)

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
