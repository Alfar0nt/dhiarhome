# Selfhosted Proxmox Dashboard
just my simple learning project for homelab monitoring, fully vibecoded btw. so if you have any bugs or error, you can simply modify it or just custom your own. Thx

Welcome to **Selfhosted Proxmox Dashboard**! 

## What is this project?
This is a simple, ultra-lightweight web dashboard designed for self-hosted home servers (like a Dell Latitude running Proxmox). It gives you a single page to monitor your server's health, see which Docker containers are running, and check if your personal websites are online.

## Purpose
Home servers often have limited resources (CPU/RAM). Many existing dashboards are heavy and require running large databases or complex setups. The purpose of this project is to provide a fast, beautiful dashboard that uses almost zero resources. 

## Tech Stacks Used
- **Backend:** Go (Golang) - incredibly fast and compiles down to a tiny, self-contained file.
- **Frontend:** HTML & Tailwind CSS - for a beautiful dark-mode design without writing tons of custom CSS.
- **Interactivity:** HTMX - allows the page to automatically update itself without writing complicated JavaScript.
- **Deployment:** Docker - making it super easy to run on any machine.

---

## How to Customize

You can customize the dashboard without changing any code! All customization happens in the `config.yaml` file.

1. Rename `config-example.yaml` to `config.yaml`.
2. Open `config.yaml` in any text editor.
3. **To add more websites to monitor:** Just add them to the `services` list at the bottom of the file.
   ```yaml
   services:
     - name: "My New Blog"
       url: "https://blog.example.com"
   ```
4. **To monitor Proxmox:** Add your Proxmox API Token ID and Secret to the `proxmox` section.
5. **To test without real data:** Set `mock: true` in the config file to see fake bouncing data while you design or test!

---

## How to Run & Compile on Your Machine

### The Easiest Way (Using Docker)
You don't need to install anything except Docker!

1. Clone this repository:
   ```bash
   git clone https://github.com/Alfar0nt/SelfHosted-Proxmox-Dashboard.git
   cd SelfHosted-Proxmox-Dashboard
   ```
2. Build the Docker image:
   ```bash
   docker build -t homelab-dash .
   ```
3. Run it! (Make sure you've created your `config.yaml` first)
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v $(pwd)/config.yaml:/app/config.yaml \
     -v /var/run/docker.sock:/var/run/docker.sock \
     homelab-dash
   ```
4. Open your browser and go to `http://localhost:8080`.

### The Developer Way (Compiling Manually)
If you want to edit the code or run it directly on your machine without Docker:

1. Install **Go** from [golang.org](https://go.dev/).
2. Clone the repository and enter the folder.
3. Build the Go application:
   ```bash
   go build -o dashboard main.go
   ```
4. Run the newly created binary:
   ```bash
   ./dashboard
   ```
5. Open your browser and go to `http://localhost:8080`.
