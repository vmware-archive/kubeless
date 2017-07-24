module.exports = {
  handler: (req, res) => {
    let body = [];
    return new Promise((resolve, reject) => {
      req.on('error', (err) => {
        reject(new Error(err));
      }).on('data', (chunk) => {
        body.push(chunk);
      }).on('end', () => {
        body = Buffer.concat(body).toString();
        console.log(body);
        res.end(body);
        resolve();
      });
    });
  },
};
