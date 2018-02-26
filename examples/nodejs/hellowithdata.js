module.exports = {
  handler: (event, context) => {
    console.log(event);
    return event.data;
  },
};
