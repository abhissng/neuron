# Neuron Package - Internal Libraries and Wrappers

This repository contains libraries and wrappers designed for use within go projects.

## Setup

Follow these steps to set up your environment and access the Neuron package:

### 1. Configure Go Environment

1.  **Set GOPRIVATE:**
    Open your terminal and execute the following command to configure the `GOPRIVATE` environment variable:

    ```bash
    go env -w GOPRIVATE="[github.com/abhissng/](https://www.google.com/search?q=https://github.com/abhissng/)*"
    ```

    This command tells Go to treat repositories under `github.com/abhissng/` as private, ensuring proper authentication.

### 2. Generate SSH Key

1.  **Generate SSH Key:**
    Generate an SSH key specifically for accessing this repository. Use the following command, replacing `<YOUR_EMAIL>` with your email address:

    ```bash
    ssh-keygen -t rsa -C "<YOUR_EMAIL>" -f ~/.ssh/github_cicd
    ```

    This will create two files: `github_cicd` (private key) and `github_cicd.pub` (public key) in your `~/.ssh/` directory.

2.  **Copy Public Key:**
    Copy the contents of the public key to your clipboard:

    ```bash
    cat ~/.ssh/github_cicd.pub | pbcopy
    ```

    You will need to add this public key to your GitHub account settings.

### 3. Configure Git

1.  **Open Git Configuration:**
    Open your global Git configuration file using a text editor:

    ```bash
    sudo nano ~/.gitconfig
    ```

2.  **Add SSH URL Configuration:**
    Add the following lines to your `~/.gitconfig` file to instruct Git to use SSH for GitHub URLs:

    ```ini
    [url "ssh://git@github.com/"]
        insteadOf = [https://github.com/](https://github.com/)
    ```

    Save and close the file.

### 4. Add GitHub to Known Hosts

1.  **Add GitHub to Known Hosts:**
    Add GitHub's SSH host key to your `known_hosts` file to prevent SSH warnings:

    ```bash
    ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
    ```

### 5. Configure SSH

1.  **Open SSH Configuration:**
    Open your SSH configuration file:

    ```bash
    sudo nano ~/.ssh/config
    ```

2.  **Add GitHub Configuration:**
    Add the following lines to your `~/.ssh/config` file, replacing `github_abhissng_rsa` with the correct name of your private key if needed:

    ```ini
    # New For github
    Host github.com
    AddKeysToAgent yes
    IdentityFile ~/.ssh/github_cicd
    ```

    Save and close the file.

### 6. Test Connection

1.  **Test SSH Connection:**
    Verify that your SSH connection to GitHub is working correctly:

    ```bash
    ssh -T git@github.com
    ```

    If the setup is correct, you should see a message confirming successful authentication.

## Usage

After completing the setup, you can import and use the libraries from this package in your Go projects:

```go
import "[github.com/abhissng/neuron/your_library](https://www.google.com/search?q=https://github.com/abhissng/neuron/your_library)"

// Use the library functions