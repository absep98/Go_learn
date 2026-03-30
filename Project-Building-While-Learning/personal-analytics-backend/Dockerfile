# =================================================================
# STAGE 1: THE BUILDER (The "Heavy" Construction Site)
# Image Size: ~300MB+ (Contains Go Compiler + Build Tools)
# =================================================================
FROM golang:1.25-alpine AS builder

# Create a workspace inside this temporary container
WORKDIR /app

# Copy only dependency files first to speed up future builds (caching)
COPY go.mod go.sum ./

# Download all external Go libraries needed for your project
RUN go mod download

# Copy your actual .go source code into the builder
COPY . .

# THE "MAGIC" BUILD COMMAND:
# 1. CGO_ENABLED=0: Tells Go NOT to use any C libraries from the OS.
#    WHY? Most Linux uses 'glibc', but Alpine uses 'musl'. If we don't
#    disable CGO, the app will crash on Alpine looking for 'glibc'.
#    This makes the binary "Static" (completely self-contained).
# 2. -o server: Names the final binary file 'server'.
# GOOS=linux    -> Explicitly target the Linux Operating System
RUN CGO_ENABLED=0  GOOS=linux go build -o server ./cmd/server


# =================================================================
# STAGE 2: THE RUNNER (The "Light" Delivery Box)
# Image Size: ~15MB (The 1GB builder is now DELETED)
# =================================================================
FROM alpine:latest

# WHY ADD THIS?
# Alpine is so tiny it doesn't even have "Trust Certificates."
# If your app connects to Redis, Databases, or APIs via HTTPS/TLS,
# it will FAIL without these certificates. We add them to the runner
# so they stay in the final image.
RUN apk add --no-cache ca-certificates

# Set the working directory in the final tiny image
WORKDIR /app

# THE HANDOFF:
# We reach back into the 'builder' stage and grab ONLY the 'server' binary.
# We DO NOT copy the .go files or the Go compiler.
# This keeps the final image tiny and secure.
COPY --from=builder /app/server /app/server

# Documentation: Tell Docker the app uses port 8080
EXPOSE 8080

# EXECUTION:
# Because we used CGO_ENABLED=0, this 'server' file is a
# "Static Binary." It carries everything it needs inside it.
# It doesn't need Go or any special C-libraries to run.
CMD ["/app/server"]
