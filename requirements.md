# Download the core Kubernetes API definition schemas
go get k8s.io/api/core/v1

# Download the client machinery library (manages connection settings and contexts)
go get k8s.io/apimachinery/pkg/apis/meta/v1

# Download the true client package that acts as the interaction engine
go get k8s.io/client-go/kubernetes

# Pull down the auth library to allow token authentication via ~/.kube/config natively
go get k8s.io/client-go/plugin/pkg/client/auth

go mod tidy


===============================

🛠️ The Fix: Force-Install Pinned Kubernetes Modules
Run this exact sequence of commands from the root directory of your project:

Bash
# 1. Pull the pinned core client module explicitly
go get k8s.io/client-go@v0.29.2

# 2. Pull the matched API machinery metadata module explicitly
go get k8s.io/apimachinery@v0.29.2

# 3. Force Go to parse the dependencies tree and update the go.sum file
go mod tidy

go get sigs.k8s.io/yaml

# Download the latest stable Kubebuilder binary
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"

# Make it executable and move it into your executable path
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

# Verify the installation was successful
kubebuilder version

# Initialize the Kubebuilder layout using the native Go plugin
kubebuilder init --domain kubemind.io --repo github.com/akshayraina999/kubemind --owner "Amit Raina"

kubebuilder create api --group v1alpha1 --version v1alpha1 --kind TriageWatcher --resource=true --controller=true

🏗️ Step 2: Generate the Native Custom Resource Definitions (CRDs)
Whenever you alter the underlying Go specification structs, you must tell Kubebuilder to compile your Go structures back out into raw Kubernetes declarative YAML schemas.

Run this command in your root project directory:

Bash
make manifests