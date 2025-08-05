package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func respond(ctx context.Context, w http.ResponseWriter, resp Encoder) error {
	// If the context has been canceled, it means the client is no longer
	// waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	statusCode := http.StatusOK

	switch resp.(type) {
	case error:
		statusCode = http.StatusInternalServerError
	default:
		if resp == nil {
			statusCode = http.StatusNoContent
		}
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := resp.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
