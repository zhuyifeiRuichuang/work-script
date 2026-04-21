"""Run with: python -m hadoop_operator (in-cluster or local kubeconfig)."""

import logging

import kopf

from hadoop_operator import controller  # noqa: F401  # registers handlers

logging.basicConfig(level=logging.INFO)

if __name__ == "__main__":
    kopf.run()
