var mongoose = require('mongoose'),
    Task = require('./api/models/todoListModel'), //created model loading here
    bodyParser = require('body-parser');

// mongoose instance connection url connection
mongoose.Promise = global.Promise;
mongoose.connect('mongodb://mongodb.default/Tododb'); 

var controller = require('./api/controllers/todoListController'); //importing route

module.exports = {
    add: function (req, res) { controller.create_a_task(req, res) },
    delete: function (req, res) { controller.delete_a_task(req, res) },
    list: function (req, res) { controller.list_all_tasks(req, res) },
    update: function (req, res) { controller.update_a_task(req, res) },
}
