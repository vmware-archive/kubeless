package de.inoio

import io.kubeless.{Context, Event}

class Handler {
  def fooBar(event: Event, context: Context): String = {
    "FOO Bar aus Hamburg"
  }
}
