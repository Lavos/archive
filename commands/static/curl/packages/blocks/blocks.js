define(['jquery'], function(jQuery){
return (function($){
	var each = function (obj, iterator) {
		for (key in obj) {
			if (obj.hasOwnProperty(key)) {
				iterator(obj[key], key, obj);
			};
		};
	};

	var extend = function (obj) {
		var objects = Array.prototype.slice.call(arguments, 1);
		var counter = 0, limit = objects.length;

		while (counter < limit) {
			var current = objects[counter];

			for (var prop in current) {
				obj[prop] = current[prop];
			};

			counter++;
		};

		return obj;
	};

	// adapted from BackBone's inherit, also adapted from Google Closure's inherit
	var ctor = function(){};

	var inherits = function (body, extend_prototypes) {
		var parent = this;

		var body = body || function(){};

		var child = function(){
			parent.apply(this, arguments);
			body.apply(this, arguments);
		};

		extend(child, parent);

		ctor.prototype = parent.prototype;
		child.prototype = new ctor();
		child.prototype.constructor = child;
		child.__super__ = parent.prototype;

		if (extend_prototypes) {
			extend(child.prototype, extend_prototypes);
		};

		return child;
	};

	// Events

	var Events = function Events (){
		this.subscriptions = {};
	};

	Events.inherits = inherits;

	var _doCallbacks = function _doCallbacks (pairs, args) {
		var safe_pairs = pairs.slice();
		var safe_args = args.slice();
		var counter = 0, limit = safe_pairs.length;

		while (counter < limit) {
			var current_pair = safe_pairs[counter];
			current_pair.callback.apply(current_pair.context, safe_args.slice(1));
			counter++;
		};
	};

	extend(Events.prototype, {
		cancelSubscriptions: function cancelSubscriptions (eventname) {
			if (eventname && this.subscriptions.hasOwnProperty(eventname)) {
				this.subscriptions[eventname] = [];
			} else if (!eventname) {
				this.subscriptions = {};
			};
		},

		once: function once (eventname, callback, context) {
			var self = this;

			function wrappedHandler () {
				callback.apply(this, arguments);
				self.off([eventname, wrappedHandler]);
			};

			return this.on(eventname, wrappedHandler, context);
		},

		on: function on (eventname, callback, context) {
			var events = eventname.split(' ');

			var event_counter = 0, event_limit = events.length;
			while (event_counter < event_limit) {
				var current_event = events[event_counter];
				this.subscriptions[current_event] = this.subscriptions[current_event] || [];

				this.subscriptions[current_event].push({
					callback: callback || function(){},
					context: context || this
				});

				event_counter++;
			};

			return callback;
		},

		// TODO more ways of removing callbacks, passing context?
		off: function off (eventname, handle) {
			if (this.subscriptions.hasOwnProperty(eventname)) {
				var callbacks = this.subscriptions[eventname];

				var counter = callbacks.length;
				while (counter--) {
					if (callbacks[counter].callback === handle) {
						this.subscriptions[eventname].splice(counter, 1);
					};
				};
			};
		},

		fire: function fire () {
			var args = Array.prototype.slice.call(arguments);
			args.splice(1, 0, this); // the firing object is always the first argument

			if (this.subscriptions.hasOwnProperty('all')) {
				_doCallbacks(this.subscriptions['all'], ['all'].concat(args));
			};

			if (this.subscriptions.hasOwnProperty(args[0])) {
				_doCallbacks(this.subscriptions[args[0]], args);
			};
		}
	});

	// TODO global events



	// Model

	var Model = Events.inherits(function Model(){
	}, {
		// static properties
		data: {},

		// methods

		get: function get (key, default_value) {
			return this.data.hasOwnProperty(key) ? this.data[key] : default_value;
		},

		set: function set (key, value, silent) {
			if (!this.hasOwnProperty('data')) {
				this.data = {};
			};

			if (this.data[key] === value) { // data is the same
				return null;
			};

			this.data[key] = value;
			var obj = {};
			obj[key] = value;

			if (!silent) {
				this.fire('change');
				this.fire('changes', obj);
			};

			return obj;
		},

		destroy: function destroy () {
			this.fire('destroy');
		},

		ingest: function ingest (data) {
			var self = this;

			if (!this.hasOwnProperty('data')) {
				this.data = {};
			};

			var changes = [];
			each(data, function(value, key){
				var change = self.set(key, value, true);

				if (change) {
					changes.push([key, value]);
				};
			});

			if (changes.length) {
				this.fire('change');
				this.fire('changes', changes);
			};
		},

		serialize: function serialize () {
			return this.data;
		}
	});



	// Collection
	// TODO more array methods

	var Collection = Events.inherits(function Collection () {}, {
		model_constructor: Model,
		length: 0,
		push: function push (){
			var args = Array.prototype.slice.call(arguments, 0);
			if (args.length) {
				Array.prototype.push.apply(this, args);
				this.fire('add', args);
				this.fire('change');

				var counter = 0, limit = args.length;
				while (counter < limit) {
					args[counter].on('all', this.event_handler, this);
					counter++;
				};
			};
		},
		event_handler: function(event_name, model){
			switch (event_name) {
			case 'destroy':
				this.remove(model);
			break;

			};
		},
		indexOf: Array.prototype.indexOf,
		splice: function(){
			var add_items = Array.prototype.slice.call(arguments, 2);
			var removed_items = Array.prototype.splice.apply(this, arguments);

			if (add_items.length) {
				this.fire('add', add_items);

				var counter = 0, limit = add_items.length;
				while (counter < limit) {
					add_items[counter].on('all', this.event_handler, this);
					counter++;
				};
			};

			if (removed_items.length) {
				this.fire('remove', removed_items);

				var counter = 0, limit = removed_items.length;
				while (counter < limit) {
					removed_items[counter].off('all', this.event_handler);
					removed_items[counter].destroy();
					counter++;
				};
			};

			if (add_items.length || removed_items.length) {
				this.fire('change');
			};

			return removed_items;
		},
		remove: function(item) {
			this.splice(this.indexOf(item), 1);
		},
		removeAll: function() {
			this.splice(0, this.length);
		},
		serialize: function() {
			var data = [], counter = 0, limit = this.length;
			while (counter < limit) {
				data.push(this[counter].data);
				counter++;
			};

			return data;
		},
		serializeToJSON: function(){
			var counter = 0, limit = this.length, working = [];
			while (counter < limit) {
				working.push(this[counter].serialize());
				counter++;
			};

			return JSON.stringify(working);
		}
	});




	// View

	var View = Events.inherits(function View (params) {
		this.element = document.createElement(this.tag_name);
		this.element.className = this.element_classes;
		this.$element = $(this.element);

		var counter = 0, limit = this.events.length;
		while (counter < limit) {
			var current_event = this.events[counter];
			this.$element.on(current_event.event_name, current_event.selector, this[current_event.function_name].bind(this));
			counter++;
		};
	}, {
		tag_name: 'div',
		element_classes: '',
		template_string: '',
		events: [],
		render: function render(data){},
		destroy: function destroy(){
			this.$element.remove();
		}
	});

	Model.inherits = Collection.inherits = View.inherits = inherits;

	var Blocks = {};
	Blocks.Model = Model;
	Blocks.Collection = Collection;
	Blocks.View = View;
	return Blocks;
})(jQuery);
});
