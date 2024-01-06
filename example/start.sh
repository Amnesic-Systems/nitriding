#!/bin/sh

nitriding -fqdn example.com &
echo "[sh] Started nitriding."

sleep 1

service.py
echo "[sh] Ran Python script."
