package servermuxoptions

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func WithAuthCookieToAuthMetadata(authCookieName, authMetadataName string) runtime.ServeMuxOption {
	return runtime.WithMetadata(func(_ context.Context, r *http.Request) metadata.MD {
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			return make(metadata.MD)
		}

		return metadata.New(map[string]string{
			authMetadataName: cookie.Value,
		})
	})
}

func WithAuthMetadataToAuthCookie(
	authMetadataName,
	authCookieName string,
	expiresInDuration time.Duration,
) runtime.ServeMuxOption {
	return runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
		md, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}

		authMetadataValues := md.HeaderMD.Get(authMetadataName)
		if len(authMetadataValues) == 0 {
			return nil
		}

		http.SetCookie(w, &http.Cookie{
			Name:     authCookieName,
			Value:    authMetadataValues[0],
			Expires:  time.Now().Add(expiresInDuration),
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			HttpOnly: true,
		})

		return nil
	})
}

func WithRemoveGoAuthMetadata(authMetadataName string) runtime.ServeMuxOption {
	return runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
		if s == authMetadataName {
			return "", false
		}
		return runtime.DefaultHeaderMatcher(s)
	})
}

func WithErrorHandler(authMetadataName string) runtime.ServeMuxOption {
	return runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		s, ok := status.FromError(err)
		if !ok {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}

		// Remove cookie
		if s.Code() == codes.Unauthenticated {
			http.SetCookie(w, &http.Cookie{
				Name:     authMetadataName,
				Value:    "",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Unix(0, 0),
				Path:     "/",
				MaxAge:   -1,
			})
		}

		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	})
}
