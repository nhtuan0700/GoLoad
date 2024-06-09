package servermuxoptions

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
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
