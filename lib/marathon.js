/******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId])
/******/ 			return installedModules[moduleId].exports;
/******/
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			exports: {},
/******/ 			id: moduleId,
/******/ 			loaded: false
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.loaded = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(0);
/******/ })
/************************************************************************/
/******/ ([
/* 0 */
/***/ function(module, exports, __webpack_require__) {

	__webpack_require__(1);
	module.exports = __webpack_require__(2);


/***/ },
/* 1 */
/***/ function(module, exports) {

	module.exports = require("babel-polyfill");

/***/ },
/* 2 */
/***/ function(module, exports, __webpack_require__) {

	/* WEBPACK VAR INJECTION */(function(module) {'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();
	
	var _commander = __webpack_require__(4);
	
	var _commander2 = _interopRequireDefault(_commander);
	
	var _start = __webpack_require__(5);
	
	var _start2 = _interopRequireDefault(_start);
	
	var _version = __webpack_require__(15);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var RootCmd = function () {
	  function RootCmd() {
	    _classCallCheck(this, RootCmd);
	
	    this.rootCmd = _commander2.default.version(_version.version);
	  }
	
	  _createClass(RootCmd, [{
	    key: 'run',
	    value: function run(args) {
	      this.rootCmd.parse(args);
	    }
	  }]);
	
	  return RootCmd;
	}();
	
	exports.default = RootCmd;
	
	
	if (!module.parent) {
	  var cmd = new RootCmd();
	  new _start2.default(cmd.rootCmd); //eslint-disable-line
	  cmd.run(process.argv);
	}
	/* WEBPACK VAR INJECTION */}.call(exports, __webpack_require__(3)(module)))

/***/ },
/* 3 */
/***/ function(module, exports) {

	module.exports = function(module) {
		if(!module.webpackPolyfill) {
			module.deprecate = function() {};
			module.paths = [];
			// module.parent = undefined by default
			module.children = [];
			module.webpackPolyfill = 1;
		}
		return module;
	}


/***/ },
/* 4 */
/***/ function(module, exports) {

	module.exports = require("commander");

/***/ },
/* 5 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _config = __webpack_require__(6);
	
	var _config2 = _interopRequireDefault(_config);
	
	var _app = __webpack_require__(7);
	
	var _app2 = _interopRequireDefault(_app);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var StartCmd = function StartCmd(rootCmd) {
	  _classCallCheck(this, StartCmd);
	
	  rootCmd.command('start').description('Start API').action(function () {
	    var app = new _app2.default(_config2.default);
	    app.run();
	  });
	};
	
	exports.default = StartCmd;

/***/ },
/* 6 */
/***/ function(module, exports) {

	module.exports = require("config");

/***/ },
/* 7 */
/***/ function(module, exports, __webpack_require__) {

	/* WEBPACK VAR INJECTION */(function(__dirname) {'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();
	
	var _humps = __webpack_require__(8);
	
	var _path = __webpack_require__(9);
	
	var _path2 = _interopRequireDefault(_path);
	
	var _koaValidate = __webpack_require__(10);
	
	var _koaValidate2 = _interopRequireDefault(_koaValidate);
	
	var _koa = __webpack_require__(11);
	
	var _koa2 = _interopRequireDefault(_koa);
	
	var _koaRoute = __webpack_require__(12);
	
	var _koaRoute2 = _interopRequireDefault(_koaRoute);
	
	var _logger = __webpack_require__(13);
	
	var _logger2 = _interopRequireDefault(_logger);
	
	var _healthcheck = __webpack_require__(17);
	
	var _healthcheck2 = _interopRequireDefault(_healthcheck);
	
	var _app = __webpack_require__(30);
	
	var _redis = __webpack_require__(18);
	
	var _postgresql = __webpack_require__(23);
	
	var _kafkaClient = __webpack_require__(27);
	
	var _kafkaProducer = __webpack_require__(29);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var MarathonApp = function () {
	  function MarathonApp(config) {
	    _classCallCheck(this, MarathonApp);
	
	    this.config = config;
	    this.allowedMethods = ['get', 'post', 'put', 'delete'];
	    this.koaApp = new _koa2.default();
	    this.configureLogger();
	    this.configureMiddleware();
	
	    this.handlersPath = _path2.default.join(__dirname, '../api/handlers');
	    this.handlers = this.getHandlers();
	    this.redisConfig = config.get('app.services.redis');
	    this.pgConfig = config.get('app.services.postgresql');
	  }
	
	  _createClass(MarathonApp, [{
	    key: 'getHandlers',
	    value: function getHandlers() {
	      var self = this;
	      var handlers = [];
	
	      // Include handlers here in the order the routes are supposed to be handled
	      // example that will be resolved correctly:
	      // GET /healthcheck/me
	      // GET /healthcheck/:param
	      // example that will be resolved incorrectly:
	      // GET /healthcheck/:param
	      // GET /healthcheck/me
	      handlers.push(new _healthcheck2.default(self));
	      handlers.push(new _app.AppsHandler(self));
	      handlers.push(new _app.AppHandler(self));
	
	      return handlers;
	    }
	  }, {
	    key: 'exit',
	    value: function exit(err) {
	      if (process.env.NODE_ENV === 'test') {
	        throw err;
	      }
	      this.logger.fatal({ err: err });
	      process.exit(1);
	    }
	  }, {
	    key: 'configureLogger',
	    value: function configureLogger() {
	      this.logger = new _logger2.default(this.config).logger.child({
	        source: 'app'
	      });
	    }
	  }, {
	    key: 'configureRedis',
	    value: function () {
	      var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee() {
	        var redisOptions;
	        return regeneratorRuntime.wrap(function _callee$(_context) {
	          while (1) {
	            switch (_context.prev = _context.next) {
	              case 0:
	                redisOptions = {
	                  db: this.redisConfig.db,
	                  shouldReconnect: this.redisConfig.shouldReconnect,
	                  password: this.redisConfig.password
	                };
	
	                if (!redisOptions.password) delete redisOptions.password;
	                _context.prev = 2;
	                _context.next = 5;
	                return (0, _redis.connect)(this.redisConfig.url, redisOptions, this.logger);
	
	              case 5:
	                this.redisClient = _context.sent;
	                _context.next = 11;
	                break;
	
	              case 8:
	                _context.prev = 8;
	                _context.t0 = _context['catch'](2);
	
	                this.exit(_context.t0);
	
	              case 11:
	              case 'end':
	                return _context.stop();
	            }
	          }
	        }, _callee, this, [[2, 8]]);
	      }));
	
	      function configureRedis() {
	        return _ref.apply(this, arguments);
	      }
	
	      return configureRedis;
	    }()
	  }, {
	    key: 'configurePostgreSQL',
	    value: function () {
	      var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee2() {
	        return regeneratorRuntime.wrap(function _callee2$(_context2) {
	          while (1) {
	            switch (_context2.prev = _context2.next) {
	              case 0:
	                _context2.prev = 0;
	                _context2.next = 3;
	                return (0, _postgresql.connect)(this.pgConfig.url, this.pgConfig.options, this.logger);
	
	              case 3:
	                this.db = _context2.sent;
	                _context2.next = 9;
	                break;
	
	              case 6:
	                _context2.prev = 6;
	                _context2.t0 = _context2['catch'](0);
	
	                this.exit(_context2.t0);
	
	              case 9:
	              case 'end':
	                return _context2.stop();
	            }
	          }
	        }, _callee2, this, [[0, 6]]);
	      }));
	
	      function configurePostgreSQL() {
	        return _ref2.apply(this, arguments);
	      }
	
	      return configurePostgreSQL;
	    }()
	  }, {
	    key: 'configureKafka',
	    value: function () {
	      var _ref3 = _asyncToGenerator(regeneratorRuntime.mark(function _callee3() {
	        var cfg, producerCfg;
	        return regeneratorRuntime.wrap(function _callee3$(_context3) {
	          while (1) {
	            switch (_context3.prev = _context3.next) {
	              case 0:
	                _context3.prev = 0;
	
	                this.logger.debug('Connecting API Kafka client...');
	                cfg = this.config.get('app.services.kafka.api.client');
	                _context3.next = 5;
	                return (0, _kafkaClient.connect)(cfg.url, cfg.clientId, this.logger);
	
	              case 5:
	                this.apiKafkaClient = _context3.sent;
	
	
	                this.logger.debug('Connecting API Kafka producer...');
	                producerCfg = this.config.get('app.services.kafka.api.producer');
	                _context3.next = 10;
	                return (0, _kafkaProducer.connect)(this.apiKafkaClient, producerCfg, this.logger);
	
	              case 10:
	                this.apiKafkaProducer = _context3.sent;
	                _context3.next = 16;
	                break;
	
	              case 13:
	                _context3.prev = 13;
	                _context3.t0 = _context3['catch'](0);
	
	                this.exit(_context3.t0);
	
	              case 16:
	              case 'end':
	                return _context3.stop();
	            }
	          }
	        }, _callee3, this, [[0, 13]]);
	      }));
	
	      function configureKafka() {
	        return _ref3.apply(this, arguments);
	      }
	
	      return configureKafka;
	    }()
	  }, {
	    key: 'initializeServices',
	    value: function () {
	      var _ref4 = _asyncToGenerator(regeneratorRuntime.mark(function _callee4() {
	        return regeneratorRuntime.wrap(function _callee4$(_context4) {
	          while (1) {
	            switch (_context4.prev = _context4.next) {
	              case 0:
	                _context4.prev = 0;
	
	                this.logger.debug('Starting redis configuration...');
	                _context4.next = 4;
	                return this.configureRedis();
	
	              case 4:
	                this.logger.debug('Starting PostgreSQL configuration...');
	                _context4.next = 7;
	                return this.configurePostgreSQL();
	
	              case 7:
	                this.logger.debug('Starting Kafka configuration...');
	                _context4.next = 10;
	                return this.configureKafka();
	
	              case 10:
	                _context4.next = 15;
	                break;
	
	              case 12:
	                _context4.prev = 12;
	                _context4.t0 = _context4['catch'](0);
	
	                this.exit(_context4.t0);
	
	              case 15:
	              case 'end':
	                return _context4.stop();
	            }
	          }
	        }, _callee4, this, [[0, 12]]);
	      }));
	
	      function initializeServices() {
	        return _ref4.apply(this, arguments);
	      }
	
	      return initializeServices;
	    }()
	  }, {
	    key: 'configureMiddleware',
	    value: function configureMiddleware() {
	      var _this = this;
	
	      this.koaApp.use(function () {
	        var _ref5 = _asyncToGenerator(regeneratorRuntime.mark(function _callee5(ctx, next) {
	          var start, ms;
	          return regeneratorRuntime.wrap(function _callee5$(_context5) {
	            while (1) {
	              switch (_context5.prev = _context5.next) {
	                case 0:
	                  start = new Date();
	                  _context5.next = 3;
	                  return next();
	
	                case 3:
	                  ms = new Date() - start;
	
	                  ctx.set('X-Response-Time', ms + 'ms');
	
	                case 5:
	                case 'end':
	                  return _context5.stop();
	              }
	            }
	          }, _callee5, _this);
	        }));
	
	        return function (_x, _x2) {
	          return _ref5.apply(this, arguments);
	        };
	      }());
	      (0, _koaValidate2.default)(this.koaApp);
	    }
	  }, {
	    key: 'initializeApp',
	    value: function () {
	      var _ref6 = _asyncToGenerator(regeneratorRuntime.mark(function _callee7() {
	        var _this2 = this;
	
	        return regeneratorRuntime.wrap(function _callee7$(_context7) {
	          while (1) {
	            switch (_context7.prev = _context7.next) {
	              case 0:
	                _context7.next = 2;
	                return this.initializeServices();
	
	              case 2:
	                this.handlers.forEach(function (handler) {
	                  _this2.allowedMethods.forEach(function (methodName) {
	                    if (!handler[methodName]) {
	                      return;
	                    }
	                    var handlerMethod = handler[methodName].bind(handler);
	                    var method = _koaRoute2.default[methodName];
	                    var args = [handler.route];
	                    var validate = handler[(0, _humps.camelize)('validate_' + methodName)];
	                    if (validate) {
	                      args.push(validate);
	                    }
	                    args.push(function () {
	                      var _ref7 = _asyncToGenerator(regeneratorRuntime.mark(function _callee6(ctx) {
	                        return regeneratorRuntime.wrap(function _callee6$(_context6) {
	                          while (1) {
	                            switch (_context6.prev = _context6.next) {
	                              case 0:
	                                _context6.next = 2;
	                                return handlerMethod(ctx);
	
	                              case 2:
	                              case 'end':
	                                return _context6.stop();
	                            }
	                          }
	                        }, _callee6, _this2);
	                      }));
	
	                      return function (_x3) {
	                        return _ref7.apply(this, arguments);
	                      };
	                    }());
	
	                    _this2.koaApp.use(method.apply(handler, args));
	                  });
	                });
	
	              case 3:
	              case 'end':
	                return _context7.stop();
	            }
	          }
	        }, _callee7, this);
	      }));
	
	      function initializeApp() {
	        return _ref6.apply(this, arguments);
	      }
	
	      return initializeApp;
	    }()
	  }, {
	    key: 'run',
	    value: function () {
	      var _ref8 = _asyncToGenerator(regeneratorRuntime.mark(function _callee8() {
	        var PORT;
	        return regeneratorRuntime.wrap(function _callee8$(_context8) {
	          while (1) {
	            switch (_context8.prev = _context8.next) {
	              case 0:
	                PORT = this.config.get('app.port');
	                _context8.next = 3;
	                return this.initializeApp();
	
	              case 3:
	
	                this.logger.info('Listening on port ' + PORT + '...');
	                this.koaApp.listen(PORT);
	
	              case 5:
	              case 'end':
	                return _context8.stop();
	            }
	          }
	        }, _callee8, this);
	      }));
	
	      function run() {
	        return _ref8.apply(this, arguments);
	      }
	
	      return run;
	    }()
	  }]);
	
	  return MarathonApp;
	}();
	
	exports.default = MarathonApp;
	/* WEBPACK VAR INJECTION */}.call(exports, "/"))

/***/ },
/* 8 */
/***/ function(module, exports) {

	module.exports = require("humps");

/***/ },
/* 9 */
/***/ function(module, exports) {

	module.exports = require("path");

/***/ },
/* 10 */
/***/ function(module, exports) {

	module.exports = require("koa-validate");

/***/ },
/* 11 */
/***/ function(module, exports) {

	module.exports = require("koa");

/***/ },
/* 12 */
/***/ function(module, exports) {

	module.exports = require("koa-route");

/***/ },
/* 13 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();
	
	var _bunyan = __webpack_require__(14);
	
	var _bunyan2 = _interopRequireDefault(_bunyan);
	
	var _version = __webpack_require__(15);
	
	var _version2 = _interopRequireDefault(_version);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var Logger = function () {
	  function Logger(config) {
	    _classCallCheck(this, Logger);
	
	    this.config = config;
	    this.logLevel = config.get('app.log.level');
	    this.logToStdOut = this.config.get('app.log.logToStdOut');
	    this.logToFile = this.config.get('app.log.logToFile');
	    this.logFile = this.config.get('app.log.file');
	    this.configureLogger();
	  }
	
	  _createClass(Logger, [{
	    key: 'getStreams',
	    value: function getStreams() {
	      var streams = [];
	      if (this.logToStdOut) {
	        streams.push({
	          stream: process.stdout,
	          level: this.logLevel
	        });
	      }
	      if (this.logToFile) {
	        streams.push({
	          path: this.logFile,
	          level: this.logLevel
	        });
	      }
	      return streams;
	    }
	  }, {
	    key: 'configureLogger',
	    value: function configureLogger() {
	      this.logger = _bunyan2.default.createLogger({
	        name: this.config.get('app.name'),
	        src: false,
	        streams: this.getStreams(),
	        serializers: { err: _bunyan2.default.stdSerializers.err }
	      }).child({ version: _version2.default }, true);
	    }
	  }]);
	
	  return Logger;
	}();
	
	exports.default = Logger;

/***/ },
/* 14 */
/***/ function(module, exports) {

	module.exports = require("bunyan");

/***/ },
/* 15 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _package = __webpack_require__(16);
	
	var _package2 = _interopRequireDefault(_package);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	var version = _package2.default.version;
	
	exports.default = version;

/***/ },
/* 16 */
/***/ function(module, exports) {

	module.exports = {
		"name": "marathon",
		"version": "0.1.0",
		"description": "Marathon push processing system",
		"main": "src/index.js",
		"scripts": {
			"test": "echo \"Error: no test specified\" && exit 1"
		},
		"repository": {
			"type": "git",
			"url": "git@git.topfreegames.com:topfreegames/marathon.git"
		},
		"author": "TFG Co",
		"license": "ISC",
		"devDependencies": {
			"babel-cli": "^6.16.0",
			"babel-core": "^6.17.0",
			"babel-istanbul": "^0.11.0",
			"babel-loader": "^6.2.5",
			"babel-plugin-syntax-async-functions": "^6.13.0",
			"babel-plugin-transform-async-to-generator": "^6.16.0",
			"babel-plugin-transform-regenerator": "^6.16.1",
			"babel-plugin-transform-runtime": "^6.15.0",
			"babel-polyfill": "^6.16.0",
			"babel-preset-es2015": "^6.16.0",
			"babel-register": "^6.16.3",
			"chai": "^3.5.0",
			"chai-http": "^3.0.0",
			"coveralls": "^2.11.14",
			"dirty-chai": "^1.2.2",
			"eslint": "^3.9.1",
			"eslint-config-airbnb-base": "^8.0.0",
			"eslint-plugin-import": "^1.16.0",
			"flow-bin": "^0.34.0",
			"graceful-fs": "^4.1.10",
			"istanbul": "^0.4.5",
			"mocha": "^3.1.2",
			"mocha-lcov-reporter": "^1.2.0",
			"plato": "^1.7.0",
			"stylish": "^1.0.0",
			"stylus": "^0.54.5",
			"supertest": "^2.0.0",
			"supertest-as-promised": "^4.0.0",
			"uuid": "^2.0.3",
			"webpack": "^1.13.2"
		},
		"dependencies": {
			"babel-runtime": "^6.11.6",
			"bluebird": "^3.4.6",
			"boom": "^4.2.0",
			"bufferutil": "^1.2.1",
			"bunyan": "^1.8.1",
			"commander": "^2.9.0",
			"config": "^1.21.0",
			"humps": "^2.0.0",
			"js-yaml": "^3.6.1",
			"json-loader": "^0.5.4",
			"kafka-node": "^1.0.5",
			"koa": "^2.0.0-alpha.7",
			"koa-route": "^3.2.0",
			"koa-validate": "^1.0.7",
			"mongoose": "^4.6.5",
			"pg": "^6.1.0",
			"pg-hstore": "^2.3.2",
			"pg-native": "^1.10.0",
			"redis": "^2.6.2",
			"redis-info": "^3.0.6",
			"redlock": "^2.0.1",
			"sequelize": "^3.24.7",
			"utf-8-validate": "^1.2.1",
			"validator": "^6.0.0"
		},
		"bin": {
			"marathon": "./lib/marathon.js"
		}
	};

/***/ },
/* 17 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();
	
	var _redis = __webpack_require__(18);
	
	var _postgresql = __webpack_require__(23);
	
	var _kafkaClient = __webpack_require__(27);
	
	var _kafkaProducer = __webpack_require__(29);
	
	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var HealthcheckHandler = function () {
	  function HealthcheckHandler(app) {
	    _classCallCheck(this, HealthcheckHandler);
	
	    this.app = app;
	    this.route = '/healthcheck';
	    this.resetServices();
	  }
	
	  _createClass(HealthcheckHandler, [{
	    key: 'resetServices',
	    value: function resetServices() {
	      this.services = {
	        redis: { up: false },
	        postgreSQL: { up: false },
	        apiKafkaClient: { up: false },
	        apiKafkaProducer: { up: false }
	      };
	    }
	  }, {
	    key: 'hasFailed',
	    value: function hasFailed() {
	      return !this.services.redis.up || !this.services.postgreSQL.up || !this.services.apiKafkaClient.up || !this.services.apiKafkaProducer.up;
	    }
	  }, {
	    key: 'get',
	    value: function () {
	      var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(ctx) {
	        return regeneratorRuntime.wrap(function _callee$(_context) {
	          while (1) {
	            switch (_context.prev = _context.next) {
	              case 0:
	                _context.next = 2;
	                return (0, _redis.check)(this.app.redisClient);
	
	              case 2:
	                this.services.redis = _context.sent;
	                _context.next = 5;
	                return (0, _postgresql.check)(this.app.db);
	
	              case 5:
	                this.services.postgreSQL = _context.sent;
	                _context.next = 8;
	                return (0, _kafkaClient.check)(this.app.apiKafkaClient);
	
	              case 8:
	                this.services.apiKafkaClient = _context.sent;
	                _context.next = 11;
	                return (0, _kafkaProducer.check)(this.app.apiKafkaProducer);
	
	              case 11:
	                this.services.apiKafkaProducer = _context.sent;
	
	                ctx.body = JSON.stringify(this.services);
	
	                if (this.hasFailed()) {
	                  ctx.status = 500;
	                }
	
	              case 14:
	              case 'end':
	                return _context.stop();
	            }
	          }
	        }, _callee, this);
	      }));
	
	      function get(_x) {
	        return _ref.apply(this, arguments);
	      }
	
	      return get;
	    }()
	  }]);
	
	  return HealthcheckHandler;
	}();
	
	exports.default = HealthcheckHandler;

/***/ },
/* 18 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	exports.withCriticalSection = exports.connect = exports.check = exports.LockTTL = exports.LockKey = undefined;
	
	var check = exports.check = function () {
	  var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(redisClient) {
	    var result, res, info;
	    return regeneratorRuntime.wrap(function _callee$(_context) {
	      while (1) {
	        switch (_context.prev = _context.next) {
	          case 0:
	            result = {
	              up: false,
	              error: null,
	              uptime: null,
	              connectedClients: null,
	              blockedClients: null,
	              usedMemory: null,
	              totalSystemMemory: null,
	              maxMemory: null,
	              rejectedConnections: 0,
	              cpuUsage: 0
	            };
	            _context.prev = 1;
	            _context.next = 4;
	            return redisClient.infoAsync();
	
	          case 4:
	            res = _context.sent;
	
	            if (res) {
	              info = _redisInfo2.default.parse(res);
	
	              result.up = true;
	              result.uptime = info.uptime_in_seconds;
	              result.connectedClients = info.connected_clients;
	              result.blockedClients = info.blocked_clients;
	              result.usedMemory = info.used_memory_human;
	              result.totalSystemMemory = info.total_system_memory_human;
	              result.maxMemory = info.maxmemory_human;
	              result.rejectedConnections = info.rejected_connections;
	              result.cpuUsage = info.used_cpu_user;
	            } else {
	              result.error = 'Could not get server status!';
	            }
	            _context.next = 11;
	            break;
	
	          case 8:
	            _context.prev = 8;
	            _context.t0 = _context['catch'](1);
	
	            result.error = _context.t0.message;
	
	          case 11:
	            return _context.abrupt('return', result);
	
	          case 12:
	          case 'end':
	            return _context.stop();
	        }
	      }
	    }, _callee, this, [[1, 8]]);
	  }));
	
	  return function check(_x) {
	    return _ref.apply(this, arguments);
	  };
	}();
	
	var connect = exports.connect = function () {
	  var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee2(redisUrl, options, logger) {
	    var logr, redisClient, hasConnected, result;
	    return regeneratorRuntime.wrap(function _callee2$(_context2) {
	      while (1) {
	        switch (_context2.prev = _context2.next) {
	          case 0:
	            logr = logger.child({
	              redisUrl: redisUrl,
	              options: options,
	              source: 'redis-extension'
	            });
	
	            logr.debug({ redisUrl: redisUrl, options: options }, 'connecting to redis...');
	            if (!options.shouldReconnect) {
	              options.retry_strategy = function () {
	                return undefined;
	              };
	            }
	            redisClient = _redis2.default.createClient(redisUrl, options);
	            hasConnected = new Promise(function (resolve, reject) {
	              if (redisClient.ready) {
	                resolve(redisClient);
	                return;
	              }
	
	              redisClient.on('ready', function () {
	                logr.debug('Connection to redis has been established successfully.');
	                resolve(redisClient);
	              });
	
	              redisClient.on('error', function (err) {
	                logr.error({ err: err }, 'Redis error connecting.');
	                reject(err);
	              });
	
	              redisClient.on('end', function () {
	                logr.error('Redis connection closed.');
	                reject(new Error('Redis connection closed.'));
	              });
	            });
	            _context2.next = 7;
	            return hasConnected;
	
	          case 7:
	            _context2.next = 9;
	            return redisClient.pingAsync();
	
	          case 9:
	            result = _context2.sent;
	
	            if (result) {
	              _context2.next = 12;
	              break;
	            }
	
	            throw new Error('Failed to get server status from redis.');
	
	          case 12:
	            logr.info({ redisUrl: redisUrl }, 'Successfully connected to redis.');
	            return _context2.abrupt('return', redisClient);
	
	          case 14:
	          case 'end':
	            return _context2.stop();
	        }
	      }
	    }, _callee2, this);
	  }));
	
	  return function connect(_x2, _x3, _x4) {
	    return _ref2.apply(this, arguments);
	  };
	}();
	
	var withCriticalSection = exports.withCriticalSection = function () {
	  var _ref3 = _asyncToGenerator(regeneratorRuntime.mark(function _callee3(redisClient, f) {
	    var options, rlock, res;
	    return regeneratorRuntime.wrap(function _callee3$(_context3) {
	      while (1) {
	        switch (_context3.prev = _context3.next) {
	          case 0:
	            if (!Lock) {
	              options = { retryCount: 50, retryDelay: 10 };
	
	              Lock = new _redlock2.default([redisClient], options);
	            }
	            _context3.next = 3;
	            return Lock.lock(LockKey, LockTTL);
	
	          case 3:
	            rlock = _context3.sent;
	            _context3.next = 6;
	            return f();
	
	          case 6:
	            res = _context3.sent;
	            _context3.next = 9;
	            return rlock.unlock();
	
	          case 9:
	            return _context3.abrupt('return', res);
	
	          case 10:
	          case 'end':
	            return _context3.stop();
	        }
	      }
	    }, _callee3, this);
	  }));
	
	  return function withCriticalSection(_x5, _x6) {
	    return _ref3.apply(this, arguments);
	  };
	}();
	
	var _redis = __webpack_require__(19);
	
	var _redis2 = _interopRequireDefault(_redis);
	
	var _bluebird = __webpack_require__(20);
	
	var _bluebird2 = _interopRequireDefault(_bluebird);
	
	var _redisInfo = __webpack_require__(21);
	
	var _redisInfo2 = _interopRequireDefault(_redisInfo);
	
	var _redlock = __webpack_require__(22);
	
	var _redlock2 = _interopRequireDefault(_redlock);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }
	
	_bluebird2.default.promisifyAll(_redis2.default.RedisClient.prototype);
	_bluebird2.default.promisifyAll(_redis2.default.Multi.prototype);
	
	var Lock = null;
	var LockKey = exports.LockKey = 'lock::matchmaking';
	var LockTTL = exports.LockTTL = 100;

/***/ },
/* 19 */
/***/ function(module, exports) {

	module.exports = require("redis");

/***/ },
/* 20 */
/***/ function(module, exports) {

	module.exports = require("bluebird");

/***/ },
/* 21 */
/***/ function(module, exports) {

	module.exports = require("redis-info");

/***/ },
/* 22 */
/***/ function(module, exports) {

	module.exports = require("redlock");

/***/ },
/* 23 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	exports.connect = exports.check = undefined;
	
	var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) { return typeof obj; } : function (obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; };
	
	var check = exports.check = function () {
	  var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(db) {
	    var result, query, deadlockQuery, res, deadlock;
	    return regeneratorRuntime.wrap(function _callee$(_context) {
	      while (1) {
	        switch (_context.prev = _context.next) {
	          case 0:
	            result = {
	              up: false
	            };
	            _context.prev = 1;
	            query = "select * from pg_stat_activity where state='active';";
	            deadlockQuery = 'SELECT blocked_locks.pid     AS blocked_pid, ' + 'blocked_locks.virtualtransaction as blocked_transaction_id, ' + 'blocked_activity.usename  AS blocked_user, ' + 'blocking_locks.pid     AS blocking_pid, ' + 'blocking_locks.virtualtransaction as blocking_transaction_id, ' + 'blocking_activity.usename AS blocking_user, ' + 'blocked_activity.query    AS blocked_statement, ' + 'blocking_activity.query   AS blocking_statement ' + 'FROM  pg_catalog.pg_locks         blocked_locks ' + 'JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid ' + 'JOIN pg_catalog.pg_locks         blocking_locks  ' + 'ON blocking_locks.locktype = blocked_locks.locktype ' + 'AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE ' + 'AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation ' + 'AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page ' + 'AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple ' + 'AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid ' + 'AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid ' + 'AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid ' + 'AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid ' + 'AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid ' + 'AND blocking_locks.pid != blocked_locks.pid ' + 'JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid ' + 'WHERE NOT blocked_locks.GRANTED; ';
	            _context.next = 6;
	            return db.client.query(query, { type: _sequelize2.default.QueryTypes.SELECT });
	
	          case 6:
	            res = _context.sent;
	
	            if (!res) {
	              _context.next = 18;
	              break;
	            }
	
	            result.activeOperations = res.length;
	            _context.next = 11;
	            return db.client.query(deadlockQuery, { type: _sequelize2.default.QueryTypes.SELECT });
	
	          case 11:
	            deadlock = _context.sent;
	
	            result.deadlock = deadlock.length === 0;
	            result.deadlockOperations = [];
	            deadlock.forEach(function (row) {
	              result.deadlockOperations.push({
	                blocked: {
	                  pid: row.blocked_pid,
	                  txId: row.blocked_transaction_id,
	                  user: row.blocked_user,
	                  statement: row.blocked_statement
	                },
	                blocking: {
	                  pid: row.blocking_pid,
	                  txId: row.blocking_transaction_id,
	                  user: row.blocking_user,
	                  statement: row.blocking_statement
	                }
	              });
	            });
	            result.up = true;
	            _context.next = 19;
	            break;
	
	          case 18:
	            result.error = 'Could not get server status!';
	
	          case 19:
	            _context.next = 24;
	            break;
	
	          case 21:
	            _context.prev = 21;
	            _context.t0 = _context['catch'](1);
	
	            result.error = _context.t0.message;
	
	          case 24:
	            return _context.abrupt('return', result);
	
	          case 25:
	          case 'end':
	            return _context.stop();
	        }
	      }
	    }, _callee, this, [[1, 21]]);
	  }));
	
	  return function check(_x) {
	    return _ref.apply(this, arguments);
	  };
	}();
	
	var connect = exports.connect = function () {
	  var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee3(pgUrl, options, logger) {
	    var _this = this;
	
	    var opt, logr, _ret;
	
	    return regeneratorRuntime.wrap(function _callee3$(_context3) {
	      while (1) {
	        switch (_context3.prev = _context3.next) {
	          case 0:
	            opt = options;
	
	            if (!options) {
	              opt = {};
	            }
	
	            logr = logger.child({
	              pgUrl: pgUrl,
	              options: options,
	              source: 'postgresql-extension'
	            });
	
	            opt.dialect = 'postgres';
	
	            logr.debug({ pgUrl: pgUrl, opt: opt }, 'Connecting to PostgreSQL...');
	
	            _context3.prev = 5;
	            return _context3.delegateYield(regeneratorRuntime.mark(function _callee2() {
	              var db, client, query, res, err;
	              return regeneratorRuntime.wrap(function _callee2$(_context2) {
	                while (1) {
	                  switch (_context2.prev = _context2.next) {
	                    case 0:
	                      db = {};
	                      client = new _sequelize2.default(pgUrl, opt);
	                      query = 'select 1;';
	                      _context2.next = 5;
	                      return client.query(query, { type: _sequelize2.default.QueryTypes.SELECT });
	
	                    case 5:
	                      res = _context2.sent;
	
	                      if (res) {
	                        _context2.next = 10;
	                        break;
	                      }
	
	                      err = new Error('Failed to connect to PostgreSQL.');
	
	                      logr.error({ pgUrl: pgUrl, err: err }, err.message);
	                      throw err;
	
	                    case 10:
	
	                      logr.debug('Loading models...');
	                      Object.keys(_index2.default).forEach(function (model) {
	                        db[model] = _index2.default[model];
	                        logr.debug({ model: model }, 'Model loaded successfully.');
	                      });
	
	                      logr.debug('All models loaded successfully.');
	
	                      logr.debug('Loading model associations...');
	                      Object.keys(db).forEach(function (modelName) {
	                        if (db[modelName].associate) {
	                          logr.debug({ modelName: modelName }, 'Loading model associations...');
	                          db[modelName].associate(db);
	                          logr.debug({ modelName: modelName }, 'Model associations loaded successfully.');
	                        }
	                      });
	
	                      db.client = client;
	                      logger.info({ pgUrl: pgUrl }, 'Successfully connected to PostgreSQL.');
	                      return _context2.abrupt('return', {
	                        v: db
	                      });
	
	                    case 18:
	                    case 'end':
	                      return _context2.stop();
	                  }
	                }
	              }, _callee2, _this);
	            })(), 't0', 7);
	
	          case 7:
	            _ret = _context3.t0;
	
	            if (!((typeof _ret === 'undefined' ? 'undefined' : _typeof(_ret)) === "object")) {
	              _context3.next = 10;
	              break;
	            }
	
	            return _context3.abrupt('return', _ret.v);
	
	          case 10:
	            _context3.next = 16;
	            break;
	
	          case 12:
	            _context3.prev = 12;
	            _context3.t1 = _context3['catch'](5);
	
	            logger.error({ pgUrl: pgUrl, err: _context3.t1 }, 'Failed to connect to PostgreSQL.');
	            throw _context3.t1;
	
	          case 16:
	          case 'end':
	            return _context3.stop();
	        }
	      }
	    }, _callee3, this, [[5, 12]]);
	  }));
	
	  return function connect(_x2, _x3, _x4) {
	    return _ref2.apply(this, arguments);
	  };
	}();
	
	var _sequelize = __webpack_require__(24);
	
	var _sequelize2 = _interopRequireDefault(_sequelize);
	
	var _index = __webpack_require__(25);
	
	var _index2 = _interopRequireDefault(_index);

	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }

/***/ },
/* 24 */
/***/ function(module, exports) {

	module.exports = require("sequelize");

/***/ },
/* 25 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _app = __webpack_require__(26);
	
	var _app2 = _interopRequireDefault(_app);
	
	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }
	
	var models = {
	  App: _app2.default
	};
	exports.default = models;

/***/ },
/* 26 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	var Sequelize = __webpack_require__(24);
	
	module.exports = function (sequelize) {
	  return sequelize.define('app', {
	    key: {
	      type: Sequelize.STRING,
	      allowNull: false,
	      validate: { len: [1, 255] }
	    },
	    bundleId: {
	      type: Sequelize.STRING,
	      allowNull: false,
	      validate: { len: [1, 2000] }
	    },
	    createdBy: {
	      type: Sequelize.STRING,
	      allowNull: false,
	      validate: { len: [1, 2000] }
	    }
	  }, {
	    timestamps: true,
	    underscored: true,
	    indexes: [{ fields: ['key'], unique: true }]
	  });
	};

/***/ },
/* 27 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	exports.connect = exports.check = undefined;
	
	var check = exports.check = function () {
	  var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(kafkaClient) {
	    var result;
	    return regeneratorRuntime.wrap(function _callee$(_context) {
	      while (1) {
	        switch (_context.prev = _context.next) {
	          case 0:
	            result = {
	              up: kafkaClient.ready,
	              error: null
	            };
	            return _context.abrupt('return', result);
	
	          case 2:
	          case 'end':
	            return _context.stop();
	        }
	      }
	    }, _callee, this);
	  }));
	
	  return function check(_x) {
	    return _ref.apply(this, arguments);
	  };
	}();
	
	var connect = exports.connect = function () {
	  var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee2(url, clientId, logger) {
	    var logr, kafkaClient, hasConnected;
	    return regeneratorRuntime.wrap(function _callee2$(_context2) {
	      while (1) {
	        switch (_context2.prev = _context2.next) {
	          case 0:
	            logr = logger.child({
	              url: url,
	              clientId: clientId,
	              source: 'kafka-client-extension'
	            });
	
	            logr.debug('Connecting to Kafka...');
	            kafkaClient = new _kafkaNode.Client(url, clientId);
	            hasConnected = new Promise(function (resolve, reject) {
	              kafkaClient.on('ready', function () {
	                logr.debug('Kafka is ready.');
	                resolve(kafkaClient);
	              });
	              kafkaClient.on('error', function (err) {
	                logr.error({ err: err }, 'Failed to connect to kafka.');
	                reject(err);
	              });
	            });
	
	
	            logr.debug('Waiting for connection...');
	            _context2.next = 7;
	            return hasConnected;
	
	          case 7:
	
	            logr.info('Successfully connected to kafka.');
	            return _context2.abrupt('return', kafkaClient);
	
	          case 9:
	          case 'end':
	            return _context2.stop();
	        }
	      }
	    }, _callee2, this);
	  }));
	
	  return function connect(_x2, _x3, _x4) {
	    return _ref2.apply(this, arguments);
	  };
	}();
	
	var _kafkaNode = __webpack_require__(28);
	
	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }

/***/ },
/* 28 */
/***/ function(module, exports) {

	module.exports = require("kafka-node");

/***/ },
/* 29 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	exports.connect = exports.check = undefined;
	
	var check = exports.check = function () {
	  var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(producer) {
	    var result;
	    return regeneratorRuntime.wrap(function _callee$(_context) {
	      while (1) {
	        switch (_context.prev = _context.next) {
	          case 0:
	            result = {
	              up: false,
	              error: null
	            };
	
	
	            try {
	              result.up = producer.ready;
	            } catch (error) {
	              result.error = error.message;
	            }
	
	            return _context.abrupt('return', result);
	
	          case 3:
	          case 'end':
	            return _context.stop();
	        }
	      }
	    }, _callee, this);
	  }));
	
	  return function check(_x) {
	    return _ref.apply(this, arguments);
	  };
	}();
	
	var connect = exports.connect = function () {
	  var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee2(client, options, logger) {
	    var logr, producer, hasConnected;
	    return regeneratorRuntime.wrap(function _callee2$(_context2) {
	      while (1) {
	        switch (_context2.prev = _context2.next) {
	          case 0:
	            logr = logger.child({
	              source: 'kafka-producer-extension',
	              options: options
	            });
	
	            logr.debug('Connecting to kafka producer...');
	            producer = new _kafkaNode2.default.Producer(client, options);
	            hasConnected = new Promise(function (resolve, reject) {
	              if (producer.ready) {
	                logr.debug('Connection to Kafka producer has been established successfully.');
	                resolve(producer);
	                return;
	              }
	
	              producer.on('ready', function () {
	                logr.debug('Connection to Kafka producer has been established successfully.');
	                resolve(producer);
	              });
	              producer.on('error', function (err) {
	                logr.error({ err: err }, 'Connection to Kafka producer failed.');
	                reject(err);
	              });
	            });
	            _context2.next = 6;
	            return hasConnected;
	
	          case 6:
	
	            logr.info('Successfully connected to Kafka producer.');
	            return _context2.abrupt('return', producer);
	
	          case 8:
	          case 'end':
	            return _context2.stop();
	        }
	      }
	    }, _callee2, this);
	  }));
	
	  return function connect(_x2, _x3, _x4) {
	    return _ref2.apply(this, arguments);
	  };
	}();
	
	var _kafkaNode = __webpack_require__(28);
	
	var _kafkaNode2 = _interopRequireDefault(_kafkaNode);

	function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }

/***/ },
/* 30 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	Object.defineProperty(exports, "__esModule", {
	  value: true
	});
	
	var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();
	
	function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }
	
	function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }
	
	var Boom = __webpack_require__(31);
	
	var AppsHandler = exports.AppsHandler = function () {
	  function AppsHandler(app) {
	    _classCallCheck(this, AppsHandler);
	
	    this.app = app;
	    this.route = '/apps';
	  }
	
	  _createClass(AppsHandler, [{
	    key: 'validatePost',
	    value: function () {
	      var _ref = _asyncToGenerator(regeneratorRuntime.mark(function _callee(ctx, next) {
	        var error;
	        return regeneratorRuntime.wrap(function _callee$(_context) {
	          while (1) {
	            switch (_context.prev = _context.next) {
	              case 0:
	                this.checkHeaders('user-email').notEmpty().isEmail();
	                this.checkQuery('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i);
	                this.checkQuery('key').notEmpty().len(1, 255);
	
	                if (!this.errors) {
	                  _context.next = 9;
	                  break;
	                }
	
	                error = Boom.badData('wrong arguments', this.errors);
	
	                ctx.status = 422;
	                ctx.body = error.output.payload;
	                ctx.body.data = error.data;
	                return _context.abrupt('return');
	
	              case 9:
	                _context.next = 11;
	                return next;
	
	              case 11:
	              case 'end':
	                return _context.stop();
	            }
	          }
	        }, _callee, this);
	      }));
	
	      function validatePost(_x, _x2) {
	        return _ref.apply(this, arguments);
	      }
	
	      return validatePost;
	    }()
	  }, {
	    key: 'get',
	    value: function () {
	      var _ref2 = _asyncToGenerator(regeneratorRuntime.mark(function _callee2(ctx) {
	        var apps;
	        return regeneratorRuntime.wrap(function _callee2$(_context2) {
	          while (1) {
	            switch (_context2.prev = _context2.next) {
	              case 0:
	                _context2.next = 2;
	                return this.app.db.App.findAll();
	
	              case 2:
	                apps = _context2.sent;
	
	                ctx.body = { apps: apps };
	                ctx.status = 200;
	
	              case 5:
	              case 'end':
	                return _context2.stop();
	            }
	          }
	        }, _callee2, this);
	      }));
	
	      function get(_x3) {
	        return _ref2.apply(this, arguments);
	      }
	
	      return get;
	    }()
	  }, {
	    key: 'post',
	    value: function () {
	      var _ref3 = _asyncToGenerator(regeneratorRuntime.mark(function _callee3(ctx) {
	        var body, app;
	        return regeneratorRuntime.wrap(function _callee3$(_context3) {
	          while (1) {
	            switch (_context3.prev = _context3.next) {
	              case 0:
	                body = this.request.body;
	
	                body.createdBy = this.request.header['user-email'];
	                _context3.next = 4;
	                return this.app.db.App.create(body);
	
	              case 4:
	                app = _context3.sent;
	
	                ctx.body = { app: app };
	                ctx.status = 201;
	
	              case 7:
	              case 'end':
	                return _context3.stop();
	            }
	          }
	        }, _callee3, this);
	      }));
	
	      function post(_x4) {
	        return _ref3.apply(this, arguments);
	      }
	
	      return post;
	    }()
	  }]);
	
	  return AppsHandler;
	}();
	
	var AppHandler = exports.AppHandler = function () {
	  function AppHandler(app) {
	    _classCallCheck(this, AppHandler);
	
	    this.app = app;
	    this.route = '/apps/:id';
	  }
	
	  _createClass(AppHandler, [{
	    key: 'validatePut',
	    value: function () {
	      var _ref4 = _asyncToGenerator(regeneratorRuntime.mark(function _callee4(ctx, next) {
	        var error;
	        return regeneratorRuntime.wrap(function _callee4$(_context4) {
	          while (1) {
	            switch (_context4.prev = _context4.next) {
	              case 0:
	                this.checkQuery('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i);
	                this.checkQuery('key').notEmpty().len(1, 255);
	
	                if (!this.errors) {
	                  _context4.next = 8;
	                  break;
	                }
	
	                error = Boom.badData('wrong arguments', this.errors);
	
	                ctx.status = 422;
	                ctx.body = error.output.payload;
	                ctx.body.data = error.data;
	                return _context4.abrupt('return');
	
	              case 8:
	                _context4.next = 10;
	                return next;
	
	              case 10:
	              case 'end':
	                return _context4.stop();
	            }
	          }
	        }, _callee4, this);
	      }));
	
	      function validatePut(_x5, _x6) {
	        return _ref4.apply(this, arguments);
	      }
	
	      return validatePut;
	    }()
	  }, {
	    key: 'get',
	    value: function () {
	      var _ref5 = _asyncToGenerator(regeneratorRuntime.mark(function _callee5(ctx) {
	        var app;
	        return regeneratorRuntime.wrap(function _callee5$(_context5) {
	          while (1) {
	            switch (_context5.prev = _context5.next) {
	              case 0:
	                _context5.next = 2;
	                return this.app.db.App.findById(this.params.id);
	
	              case 2:
	                app = _context5.sent;
	
	                if (app) {
	                  _context5.next = 6;
	                  break;
	                }
	
	                ctx.status = 404;
	                return _context5.abrupt('return');
	
	              case 6:
	                ctx.body = { app: app };
	                ctx.status = 200;
	
	              case 8:
	              case 'end':
	                return _context5.stop();
	            }
	          }
	        }, _callee5, this);
	      }));
	
	      function get(_x7) {
	        return _ref5.apply(this, arguments);
	      }
	
	      return get;
	    }()
	  }, {
	    key: 'put',
	    value: function () {
	      var _ref6 = _asyncToGenerator(regeneratorRuntime.mark(function _callee6(ctx) {
	        var body, app, updatedApp;
	        return regeneratorRuntime.wrap(function _callee6$(_context6) {
	          while (1) {
	            switch (_context6.prev = _context6.next) {
	              case 0:
	                body = this.request.body;
	                _context6.next = 3;
	                return this.app.db.App.findById(this.params.id);
	
	              case 3:
	                app = _context6.sent;
	
	                if (app) {
	                  _context6.next = 7;
	                  break;
	                }
	
	                ctx.status = 404;
	                return _context6.abrupt('return');
	
	              case 7:
	                _context6.next = 9;
	                return app.updateAttributes(body);
	
	              case 9:
	                updatedApp = _context6.sent;
	
	                ctx.body = { app: updatedApp };
	                ctx.status = 200;
	
	              case 12:
	              case 'end':
	                return _context6.stop();
	            }
	          }
	        }, _callee6, this);
	      }));
	
	      function put(_x8) {
	        return _ref6.apply(this, arguments);
	      }
	
	      return put;
	    }()
	  }, {
	    key: 'delete',
	    value: function () {
	      var _ref7 = _asyncToGenerator(regeneratorRuntime.mark(function _callee7(ctx) {
	        var app;
	        return regeneratorRuntime.wrap(function _callee7$(_context7) {
	          while (1) {
	            switch (_context7.prev = _context7.next) {
	              case 0:
	                _context7.next = 2;
	                return this.app.db.App.findById(this.params.id);
	
	              case 2:
	                app = _context7.sent;
	
	                if (app) {
	                  _context7.next = 6;
	                  break;
	                }
	
	                ctx.status = 404;
	                return _context7.abrupt('return');
	
	              case 6:
	                _context7.next = 8;
	                return app.destroy();
	
	              case 8:
	                ctx.status = 204;
	
	              case 9:
	              case 'end':
	                return _context7.stop();
	            }
	          }
	        }, _callee7, this);
	      }));

	      function _delete(_x9) {
	        return _ref7.apply(this, arguments);
	      }

	      return _delete;
	    }()
	  }]);

	  return AppHandler;
	}();

/***/ },
/* 31 */
/***/ function(module, exports) {

	module.exports = require("boom");

/***/ }
/******/ ]);
//# sourceMappingURL=marathon.js.map