
# kubectl-multi

A lightweight kubectl plugin that allows you to query resources across multiple Kubernetes clusters simultaneously.

## 🌐 Overview

`kubectl-multi` is designed for developers working with multi-cluster environments — such as those set up with KubeStellar, OCM, or Kind. It fetches and displays resources (like pods, nodes, etc.) across all clusters defined in your kubeconfig contexts, making debugging, observation, and testing seamless.

## ✅ Features

### 🔍 Unified Multi-Cluster Queries
Retrieve resource data across multiple clusters using one command:

```bash
kubectl multi get pods
```

### 💡 Intelligent Kubeconfig Discovery
* Tries to load kubeconfigs for all ManagedCluster resources.
* Falls back to known `kubectl` contexts if no custom path is specified.
* Automatically manages and caches discovered clusters in a `clustermap.json` file.

### 🧠 Resource Scope Detection
* Differentiates between cluster-scoped (e.g., nodes, namespaces) and namespace-scoped (e.g., pods, services) resources.
* Allows specifying namespaces with the `--namespace` flag.

### 📦 Output Compatibility
* Outputs clean raw JSON for namespace-scoped resources.
* Lists names for cluster-scoped resources.
* Designed to work cleanly with piping tools like `jq`.

## 🔧 Installation

### 1. Clone & Build
```bash
git clone [https://github.com/your-username/kubectl-multi.git](https://github.com/your-username/kubectl-multi.git)
cd kubectl-multi
go build -o kubectl-multi
```

### 2. Install the Plugin
```bash
sudo mv kubectl-multi /usr/local/bin/
```

### 3. Verify
```bash
kubectl multi get nodes
```

## 🛠️ Usage

```bash
kubectl multi get <resource> [--namespace <ns>]
```

### Examples
```bash
kubectl multi get pods
kubectl multi get services --namespace kube-system
kubectl multi get nodes
```

## 📁 Internals

### How It Works

**Local Cluster**
* Uses the default `kubectl` binary to fetch resources from the current context.

**Remote Clusters**
* Uses the OCM `ManagedCluster` CRD to discover remote cluster names.
* Attempts to resolve a kubeconfig for each:
    * From `clustermap.json`
    * From local context match (`kubectl config get-contexts`)
    * From Kind fallback path: `~/.kube/kind-config-<cluster-name>`

**Kubeconfig Resolution**
* If no kubeconfig is found, the cluster is skipped with a warning.
* All valid kubeconfigs are used to connect and fetch resources dynamically.

**Resource Handling**
* **Cluster-scoped:** uses dynamic client and GVR discovery.
* **Namespace-scoped:** uses core client and `.DoRaw()` to pull the resource list and format JSON.

## 📂 File Structure

```text
kubectl-multi/
├── main.go
└── clustermap.json     # auto-generated by plugin
```

## 🧪 Requirements

* Go ≥ 1.20
* A valid kubeconfig with access to multiple clusters
* Kubernetes clusters reachable from your machine (e.g., via Kind or KubeStellar)

## 🧱 Roadmap

* [ ] Support `apply`, `delete`, and `describe`
* [ ] Output table summaries by default (opt-in raw JSON)
* [ ] Group by resource instead of cluster (reverse layout)
* [ ] Filter clusters (e.g., `--only kind-*`)
* [ ] Add color-coded status indicators

## 🧼 Cleanup

To remove the plugin:

```bash
sudo rm /usr/local/bin/kubectl-multi
```

And optionally delete:

```bash
rm clustermap.json
```
```
