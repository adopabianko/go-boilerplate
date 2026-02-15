package request

import "github.com/gin-gonic/gin"

// GetTimeLocation returns the timezone location from the X-Timezone header.
// If the header is not provided, it defaults to "UTC".
func GetTimeLocation(c *gin.Context) string {
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		return "UTC"
	}
	return tz
}
