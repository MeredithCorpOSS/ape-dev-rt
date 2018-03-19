---
layout: "null"
page_title: "Provider: Null"
sidebar_current: "docs-null-index"
description: |-
  The null provider provides no-op constructs that can be useful helpers in tricky cases.
---

# Null Provider

The `null` provider is a rather-unusual provider that has constructs that
intentionally do nothing. This may sound strange, and indeed these constructs
do not need to be used in most cases, but they can be useful in various
situations to help orchestrate tricky behavior or work around limitations.

The documentation of each feature of this provider, accessible via the
navigation, gives examples of situations where these constructs may prove
useful.

Usage of the `null` provider can make a Terraform configuration harder to
understand. While it can be useful in certain cases, it should be applied with
care and other solutions preferred when available.
