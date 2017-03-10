module.exports = {
  handler: function (req, res) {
    console.log(req.body)
    res.end('hello world')
  }
}
