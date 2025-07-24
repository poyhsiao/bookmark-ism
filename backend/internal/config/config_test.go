package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad tests the configuration loading functionality
// TestLoad 測試配置載入功能
func TestLoad(t *testing.T) {
	t.Run("Load with Default Values", func(t *testing.T) {
		// Clear environment variables
		// 清除環境變量
		clearEnvVars()

		config, err := Load()
		require.NoError(t, err)

		// Test default values
		// 測試默認值
		assert.Equal(t, "8080", config.Server.Port)
		assert.Equal(t, "0.0.0.0", config.Server.Host)
		assert.Equal(t, "development", config.Server.Environment)
		assert.Equal(t, 30, config.Server.ReadTimeout)
		assert.Equal(t, 30, config.Server.WriteTimeout)

		assert.Equal(t, "localhost", config.Database.Host)
		assert.Equal(t, "5432", config.Database.Port)
		assert.Equal(t, "postgres", config.Database.User)
		assert.Equal(t, "postgres", config.Database.Password)
		assert.Equal(t, "bookmark_sync", config.Database.DBName)
		assert.Equal(t, "disable", config.Database.SSLMode)
		assert.Equal(t, 25, config.Database.MaxConns)
		assert.Equal(t, 5, config.Database.MinConns)

		assert.Equal(t, "localhost", config.Redis.Host)
		assert.Equal(t, "6379", config.Redis.Port)
		assert.Equal(t, "", config.Redis.Password)
		assert.Equal(t, 0, config.Redis.DB)
		assert.Equal(t, 10, config.Redis.PoolSize)

		assert.Equal(t, "http://localhost:8000", config.Supabase.URL)
		assert.Equal(t, "", config.Supabase.AnonKey)
		assert.Equal(t, "http://localhost:9999", config.Supabase.AuthURL)
		assert.Equal(t, "ws://localhost:4000", config.Supabase.RealtimeURL)

		assert.Equal(t, "localhost:9000", config.Storage.Endpoint)
		assert.Equal(t, "minioadmin", config.Storage.AccessKeyID)
		assert.Equal(t, "minioadmin", config.Storage.SecretAccessKey)
		assert.Equal(t, "bookmarks", config.Storage.BucketName)
		assert.False(t, config.Storage.UseSSL)

		assert.Equal(t, "localhost", config.Search.Host)
		assert.Equal(t, "8108", config.Search.Port)
		assert.Equal(t, "xyz", config.Search.APIKey)

		assert.Equal(t, "your-secret-key", config.JWT.Secret)
		assert.Equal(t, 24, config.JWT.ExpiryHour)

		assert.Equal(t, "info", config.Logger.Level)
		assert.Equal(t, "json", config.Logger.Format)
		assert.Equal(t, "stdout", config.Logger.OutputPath)
	})

	t.Run("Load with Environment Variables", func(t *testing.T) {
		// Set environment variables
		// 設置環境變量
		envVars := map[string]string{
			"SERVER_PORT":        "9090",
			"SERVER_HOST":        "127.0.0.1",
			"SERVER_ENVIRONMENT": "production",
			"DATABASE_HOST":      "db.example.com",
			"DATABASE_PORT":      "5433",
			"DATABASE_USER":      "dbuser",
			"DATABASE_PASSWORD":  "dbpass",
			"DATABASE_DBNAME":    "mydb",
			"DATABASE_SSLMODE":   "require",
			"DATABASE_MAX_CONNS": "50",
			"DATABASE_MIN_CONNS": "10",
			"REDIS_HOST":         "redis.example.com",
			"REDIS_PORT":         "6380",
			"REDIS_PASSWORD":     "redispass",
			"REDIS_DB":           "1",
			"REDIS_POOL_SIZE":    "20",
			"JWT_SECRET":         "my-secret-key",
			"JWT_EXPIRY_HOUR":    "48",
			"LOGGER_LEVEL":       "debug",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
		}
		defer clearEnvVars()

		config, err := Load()
		require.NoError(t, err)

		// Test environment variable values
		// 測試環境變量值
		assert.Equal(t, "9090", config.Server.Port)
		assert.Equal(t, "127.0.0.1", config.Server.Host)
		assert.Equal(t, "production", config.Server.Environment)

		assert.Equal(t, "db.example.com", config.Database.Host)
		assert.Equal(t, "5433", config.Database.Port)
		assert.Equal(t, "dbuser", config.Database.User)
		assert.Equal(t, "dbpass", config.Database.Password)
		assert.Equal(t, "mydb", config.Database.DBName)
		assert.Equal(t, "require", config.Database.SSLMode)
		assert.Equal(t, 50, config.Database.MaxConns)
		assert.Equal(t, 10, config.Database.MinConns)

		assert.Equal(t, "redis.example.com", config.Redis.Host)
		assert.Equal(t, "6380", config.Redis.Port)
		assert.Equal(t, "redispass", config.Redis.Password)
		assert.Equal(t, 1, config.Redis.DB)
		assert.Equal(t, 20, config.Redis.PoolSize)

		assert.Equal(t, "my-secret-key", config.JWT.Secret)
		assert.Equal(t, 48, config.JWT.ExpiryHour)

		assert.Equal(t, "debug", config.Logger.Level)
	})

	t.Run("Load with Mixed Sources", func(t *testing.T) {
		// Set some environment variables
		// 設置一些環境變量
		os.Setenv("SERVER_PORT", "8888")
		os.Setenv("DATABASE_HOST", "custom.db.com")
		defer clearEnvVars()

		config, err := Load()
		require.NoError(t, err)

		// Environment variables should override defaults
		// 環境變量應該覆蓋默認值
		assert.Equal(t, "8888", config.Server.Port)
		assert.Equal(t, "custom.db.com", config.Database.Host)

		// Other values should remain default
		// 其他值應該保持默認
		assert.Equal(t, "0.0.0.0", config.Server.Host)
		assert.Equal(t, "5432", config.Database.Port)
	})

	t.Run("Load with Invalid Integer Values", func(t *testing.T) {
		// Set invalid integer environment variable
		// 設置無效的整數環境變量
		os.Setenv("SERVER_READ_TIMEOUT", "invalid")
		defer clearEnvVars()

		// Should still load with default value
		// 應該仍然使用默認值載入
		config, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 30, config.Server.ReadTimeout) // Default value
	})
}

// TestServerConfig tests the ServerConfig structure
// TestServerConfig 測試 ServerConfig 結構
func TestServerConfig(t *testing.T) {
	config := ServerConfig{
		Port:         "8080",
		Host:         "localhost",
		ReadTimeout:  30,
		WriteTimeout: 30,
		Environment:  "development",
	}

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 30, config.ReadTimeout)
	assert.Equal(t, 30, config.WriteTimeout)
	assert.Equal(t, "development", config.Environment)
}

// TestDatabaseConfig tests the DatabaseConfig structure
// TestDatabaseConfig 測試 DatabaseConfig 結構
func TestDatabaseConfig(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		SSLMode:  "disable",
		MaxConns: 25,
		MinConns: 5,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "postgres", config.User)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "testdb", config.DBName)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxConns)
	assert.Equal(t, 5, config.MinConns)
}

// TestRedisConfig tests the RedisConfig structure
// TestRedisConfig 測試 RedisConfig 結構
func TestRedisConfig(t *testing.T) {
	config := RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "password",
		DB:       0,
		PoolSize: 10,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "6379", config.Port)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
}

// TestSupabaseConfig tests the SupabaseConfig structure
// TestSupabaseConfig 測試 SupabaseConfig 結構
func TestSupabaseConfig(t *testing.T) {
	config := SupabaseConfig{
		URL:         "http://localhost:8000",
		AnonKey:     "anon-key",
		AuthURL:     "http://localhost:9999",
		RealtimeURL: "ws://localhost:4000",
	}

	assert.Equal(t, "http://localhost:8000", config.URL)
	assert.Equal(t, "anon-key", config.AnonKey)
	assert.Equal(t, "http://localhost:9999", config.AuthURL)
	assert.Equal(t, "ws://localhost:4000", config.RealtimeURL)
}

// TestStorageConfig tests the StorageConfig structure
// TestStorageConfig 測試 StorageConfig 結構
func TestStorageConfig(t *testing.T) {
	config := StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
		BucketName:      "test-bucket",
		UseSSL:          true,
	}

	assert.Equal(t, "localhost:9000", config.Endpoint)
	assert.Equal(t, "access-key", config.AccessKeyID)
	assert.Equal(t, "secret-key", config.SecretAccessKey)
	assert.Equal(t, "test-bucket", config.BucketName)
	assert.True(t, config.UseSSL)
}

// TestSearchConfig tests the SearchConfig structure
// TestSearchConfig 測試 SearchConfig 結構
func TestSearchConfig(t *testing.T) {
	config := SearchConfig{
		Host:   "localhost",
		Port:   "8108",
		APIKey: "api-key",
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "8108", config.Port)
	assert.Equal(t, "api-key", config.APIKey)
}

// TestJWTConfig tests the JWTConfig structure
// TestJWTConfig 測試 JWTConfig 結構
func TestJWTConfig(t *testing.T) {
	config := JWTConfig{
		Secret:     "secret-key",
		ExpiryHour: 24,
	}

	assert.Equal(t, "secret-key", config.Secret)
	assert.Equal(t, 24, config.ExpiryHour)
}

// TestLoggerConfig tests the LoggerConfig structure
// TestLoggerConfig 測試 LoggerConfig 結構
func TestLoggerConfig(t *testing.T) {
	config := LoggerConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}

	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "stdout", config.OutputPath)
}

// clearEnvVars clears all environment variables used in tests
// clearEnvVars 清除測試中使用的所有環境變量
func clearEnvVars() {
	envVars := []string{
		"SERVER_PORT", "SERVER_HOST", "SERVER_ENVIRONMENT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER", "DATABASE_PASSWORD", "DATABASE_DBNAME",
		"DATABASE_SSLMODE", "DATABASE_MAX_CONNS", "DATABASE_MIN_CONNS",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB", "REDIS_POOL_SIZE",
		"SUPABASE_URL", "SUPABASE_ANON_KEY", "SUPABASE_AUTH_URL", "SUPABASE_REALTIME_URL",
		"STORAGE_ENDPOINT", "STORAGE_ACCESS_KEY_ID", "STORAGE_SECRET_ACCESS_KEY", "STORAGE_BUCKET_NAME", "STORAGE_USE_SSL",
		"SEARCH_HOST", "SEARCH_PORT", "SEARCH_API_KEY",
		"JWT_SECRET", "JWT_EXPIRY_HOUR",
		"LOGGER_LEVEL", "LOGGER_FORMAT", "LOGGER_OUTPUT_PATH",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
