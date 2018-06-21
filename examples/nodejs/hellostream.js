const from = require('from2');
const eos = require('end-of-stream');

function fromString(string) {
    return from(function(size, next) {
        if (string.length <= 0) return next(null, null);

        const chunk = string.slice(0, size);
        string = string.slice(size);

        next(null, chunk);
  });
}

module.exports = {
    foo: function(event, context) {
        return new Promise((resolve, reject) => {
            const {response} = event.extensions;
            const stream = fromString('hello world!');

            eos(stream, err => err ? reject(err) : resolve(stream));

            response.setHeader('Content-Type', 'text/event-stream; charset=utf-8');
            stream.pipe(response);
        });
    }
}
