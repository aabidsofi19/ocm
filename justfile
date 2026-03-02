set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

default:
  @just --list

# Run the analyzer + server (dashboard at http://HOST:PORT/)
run REPO="../meshery/" DB="ocm.sqlite" HOST="127.0.0.1" PORT="8080" SERVICE_KEY="manifest":
  go run ./cmd/ocm --repo "{{REPO}}" --db "{{DB}}" --host "{{HOST}}" --port "{{PORT}}" --service-key "{{SERVICE_KEY}}"

test:
  go test ./...

tidy:
  go mod tidy

build OUT="ocm":
  go build -o "{{OUT}}" ./cmd/ocm

phase2-pdf:
  typst compile docs/PHASE_2.typ specs/phase-2.pdf
