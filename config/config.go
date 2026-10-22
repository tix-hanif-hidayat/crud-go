package config

import "time"

var REFRESH_EXCHANGE_RATE_INTERVAL = 5 * time.Second
var LOCAL_EXCHANGE_RATE_EXPIRATION = 5 * time.Second
var EXCHANGE_RATE_API_URL = "https://api.frankfurter.dev/v2/rates?base=idr"
