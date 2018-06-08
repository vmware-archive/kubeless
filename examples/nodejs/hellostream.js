var from = require('from2')

function fromString(string) {
  return from(function(size, next) {
    if (string.length <= 0) return next(null, null)

    var chunk = string.slice(0, size)
    string = string.slice(size)

    next(null, chunk)
  })
}

module.exports = {
  foo: function (event, context) {
    return fromString('hello world!')
  }
}
