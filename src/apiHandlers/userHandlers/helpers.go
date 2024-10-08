package userHandlers

import (
	"ExamSphere/src/apiHandlers"
	"ExamSphere/src/core/appConfig"
	"ExamSphere/src/core/utils/hashing"
	"ExamSphere/src/database"
	"encoding/hex"

	"strings"
	"sync"
	"time"

	"github.com/ALiwoto/ssg/ssg"
	"github.com/gofiber/fiber/v2"
	fUtils "github.com/gofiber/fiber/v2/utils"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(userInfo *database.UserInfo) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userInfo.UserId
	claims["refresh"] = false
	claims["auth_hash"] = userInfo.AuthHash
	claims["exp"] = time.Now().Add(appConfig.AccessTokenExpiration).Unix()
	accessToken, _ := token.SignedString(appConfig.AccessTokenSigningKey)
	return accessToken
}

func GenerateRefreshToken(userInfo *database.UserInfo) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userInfo.UserId
	claims["refresh"] = true
	claims["auth_hash"] = userInfo.AuthHash
	claims["exp"] = time.Now().Add(appConfig.RefreshTokenExpiration).Unix()
	refreshToken, _ := token.SignedString(appConfig.RefreshTokenSigningKey)
	return refreshToken
}

// getLoginExpiration returns the expiration time of the login
// done on the master server. Do note that this will send the
// little bit less, so that the client can refresh the token
// before it expires.
func getLoginExpiration() int64 {
	return time.Now().Add(
		appConfig.AccessTokenExpiration - time.Hour,
	).Unix()
}

func IsInvalidPassword(value string) bool {
	return len(value) < MinPasswordLength ||
		len(value) > MaxPasswordLength
}

func IsEmailValid(email string) bool {
	// minimum length of an email is 3
	// e.g. a@a (if the domain is 1 character)
	return len(email) >= 3 && emailRegex.MatchString(email)
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return apiHandlers.SendErrMalformedJWT(c)
	}

	return apiHandlers.SendErrInvalidJWT(c)
}

func isRateLimited(c *fiber.Ctx) bool {
	path := strings.ToLower(fUtils.CopyString(c.Path()))
	entryKey := strings.ToLower(fUtils.CopyString(c.IP())) + "_" +
		path

	entryValue := requestRateLimitMap.Get(entryKey)
	if entryValue == nil {
		entryValue = &userRequestEntry{
			RequestPath: path,
			LastTryAt:   time.Now(),
			TryCount:    1,
			mut:         &sync.Mutex{},
		}
		requestRateLimitMap.Add(entryKey, entryValue)
		return false
	}

	entryValue.mut.Lock()
	defer entryValue.mut.Unlock()

	// check time
	if time.Since(entryValue.LastTryAt) > appConfig.GetMaxRateLimitDuration() {
		entryValue.TryCount = 1
		entryValue.LastTryAt = time.Now()
		return false
	}

	// is already rate limited?
	if entryValue.TryCount > appConfig.GetMaxRequestTillRateLimit() {
		if time.Since(entryValue.LastTryAt) > appConfig.GetRateLimitPunishmentDuration() {
			// it should get released now
			entryValue.TryCount = 1
			entryValue.LastTryAt = time.Now()
			return false
		}
		return true
	}

	entryValue.TryCount++
	if entryValue.TryCount > appConfig.GetMaxRequestTillRateLimit() {
		entryValue.LastTryAt = time.Now()
		return true
	}

	return false
}

func toSearchedUsersResult(users []*database.UserInfo) []SearchedUserInfo {
	searchedUsers := make([]SearchedUserInfo, 0, len(users))

	for _, user := range users {
		searchedUsers = append(searchedUsers, SearchedUserInfo{
			UserId:    user.UserId,
			FullName:  user.FullName,
			Role:      user.Role,
			Email:     user.Email,
			IsBanned:  user.IsBanned,
			BanReason: user.BanReason,
			CreatedAt: user.CreatedAt,
		})
	}

	return searchedUsers
}

func newChangePasswordRequest(userInfo *database.UserInfo) (*changePasswordRequestEntry, error) {
	entry := changePasswordRequestMap.Get(userInfo.UserId)

	if entry != nil {
		entry.mut.Lock()
		defer entry.mut.Unlock()

		// the user has previous attempt in our current time-frame.
		if entry.TryCount >= MaxPasswordRequestAttempts {
			return nil, ErrTooManyPasswordChangeAttempts
		} else if time.Since(entry.LastTryAt) < MinPasswordAttemptWaitTime {
			return nil, ErrTooManyPasswordChangeAttempts
		}

		entry.TryCount++
		entry.LastTryAt = time.Now()

		// since we are going to generate new RqId parameter, we should delete the old one
		changePasswordRequestMap.Delete(reqFirst + entry.RqId)
	} else {
		entry = &changePasswordRequestEntry{
			UserId:    userInfo.UserId,
			mut:       &sync.Mutex{},
			LastTryAt: time.Now(),
			TryCount:  1,
		}
	}

	// generate the request parameters
	entry.RqId = fUtils.UUIDv4()
	entry.LTNum = passwordChangeRqGenerator.Next()
	entry.RTParam = hex.EncodeToString([]byte(
		userInfo.AuthHash + entry.RqId + ssg.ToBase10(entry.LTNum)))

	changePasswordRequestMap.Add(userInfo.UserId, entry)
	changePasswordRequestMap.Add(reqFirst+entry.RqId, entry)

	return entry, nil
}

func newConfirmAccountRequest(userInfo *database.UserInfo) (*confirmAccountRequestEntry, error) {
	if userInfo.SetupCompleted {
		return nil, ErrAccountAlreadyConfirmed
	}

	entry := &confirmAccountRequestEntry{
		UserId:       userInfo.UserId,
		ConfirmToken: hashing.HashSHA256(userInfo.AuthHash + userInfo.Email),
		RLToken:      hashing.HashSHA512(userInfo.AuthHash + "_" + userInfo.Email),
	}

	return entry, nil
}

func verifyAccountConfirmation(userInfo *database.UserInfo, data *ConfirmAccountData) bool {
	return userInfo.UserId == data.UserId &&
		hashing.CompareSHA256(data.ConfirmToken, userInfo.AuthHash+userInfo.Email) &&
		hashing.CompareSHA512(data.RLToken, userInfo.AuthHash+"_"+userInfo.Email)
}

func getChangePasswordRequest(rqId string) *changePasswordRequestEntry {
	entry := changePasswordRequestMap.Get(reqFirst + rqId)
	if entry == nil {
		return nil
	}

	return entry
}
