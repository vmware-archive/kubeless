module.exports = {
  handler: function (req, res) {
    var body = []
    req.on('error', function (err) {
      console.error(err)
    }).on('data', function (chunk) {
      body.push(chunk)
    }).on('end', function () {
      body = Buffer.concat(body).toString()
      console.log(body)
      res.end(body)
    })
  }
}
