package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/services"
)

func CreateUploadHandler(us *services.UploadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, fh, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "error opening file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		generateThumbnail := r.FormValue("generateThumbnail")
		thumbnailPosition, posErr := strconv.ParseFloat(r.FormValue("thumbnailPosition"), 64)

		input := services.UploadFileInput{
			File:              file,
			FileHeader:        fh,
			GenerateThumbnail: null.NewBool(true, generateThumbnail == "true").Ptr(),
			ThumbnailPosition: null.NewFloat(thumbnailPosition, posErr == nil).Ptr(),
		}

		result, err := us.UploadFile(r.Context(), input)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "upload failed", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}
