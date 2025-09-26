# Dockerfile for Kubeflow ModelKit Component
FROM python:3.10-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install ModelKit CLI (replace with actual install if available via pip or other means)
# Example: pip install modelkit-cli
# If ModelKit is only available via binary, add COPY or curl/wget here
# For now, placeholder:
RUN pip install --no-cache-dir ml-metadata kfp
# TODO: Add ModelKit CLI installation here

# Copy component code
COPY component.py /component.py

ENTRYPOINT ["python", "/component.py"]
