FROM golang:1.24.4-bookworm

# Install dependencies
RUN apt-get update && apt-get install -y \
    git sqlite3 make curl sudo \
		ca-certificates gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install NodeJS 22.x using the offical NodeSource setup
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub Copilot CLI globally
RUN npm install -g @github/copilot

# Create non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Pre-create the .copilot directory with correct ownership
# This ensures that when the named volume is first created, it inherits these permissions.
RUN mkdir -p /home/vscode/.copilot && \
    chown -R vscode:vscode /home/vscode/.copilot

USER $USERNAME
WORKDIR /workspace

# Add pm binary to PATH (points to /workspace/bin/pm)
ENV PATH="/workspace/bin:${PATH}"
