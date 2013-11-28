define([
	'jquery', 'underscore', 'doubleunderscore', 'blocks',
	'text!./main.jst', 'text!./list.jst', 'text!./item.jst', 'text!./editor.jst'
], function(
	$, _, __, Blocks,
	main_template_string, list_template_string, item_template_string, editor_template_string
){
	var Note = Blocks.Model.inherits(function(){
		this.data = {
			hex: '',
			title: '',
			content: ''
		};
	}, {});

	var Notes = Blocks.Collection.inherits(function(){

	}, {});


	var List = Blocks.View.inherits(function(inf){
		this.collection = new Notes();
		this.collection.on('change', this.render, this);

		this.inf = inf;

		this.element.innerHTML = list_template_string;
		this.target = this.$element.find('.item_target')[0];
		this.$search = this.$element.find('.search');
	}, {
		tag_name: 'nav',
		element_classes: 'list',

		events: [
			{ event_name: 'input', selector: '.search', function_name: 'search' },
		],

		render: function(){
			var frag = document.createDocumentFragment();
			var counter = 0, limit = this.collection.length;
			while (counter < limit) {
				var item = new Item(this.collection[counter]);
				frag.appendChild(item.element);

				item.on('select', this.select, this);
				counter++;
			};

			this.target.innerHTML = '';
			this.target.appendChild(frag);
		},

		search: function(){
			var self = this;
	
			$.ajax({
				type: 'GET',
				url: '/search',
				dataType: 'json',
				data: { q: this.$search.val() },
				success: function (data) {
					self.collection.removeAll();

					var counter = 0, limit = data.length;
					while (counter < limit) {
						var note = new Note();
						note.ingest(data[counter]);
						self.collection.push(note);
						counter++;
					};
				
				}
			});
		},

		select: function(item, model){
			this.fire('select', model);
		}		
	});

	var Item = Blocks.View.inherits(function(model){
		this.model = model;
		this.element.innerHTML = this.comp_template(this.model.data);
	}, {
		tag_name: 'article',
		comp_template: __.template(item_template_string),
		events: [
			{ event_name: 'click', selector: '*', function_name: 'select' },
		],

		select: function(){
			this.fire('select', this.model);
		}
	});

	var Editor = Blocks.View.inherits(function(){
		this.element.innerHTML = this.comp_template(this.model);
		this.$textarea = this.$element.find('textarea.content');
		this.$title = this.$element.find('input.title');
	}, {
		tag_name: 'section',
		element_classes: 'editor',
		comp_template: __.template(editor_template_string),

		events: [
			{ event_name: 'change', selector: 'textarea', function_name: 'update' },
			{ event_name: 'click', selector: '[data-action="save"]', function_name: 'save' },
		],

		render: function(){
		},

		edit: function(item, model){
			var self = this;

			self.model = model;
			console.dir(self.model);

			var rr = self.model.get('revision_refs');
			var last = rr[rr.length-1];

			$.ajax({
				type: 'GET',
				url: last,
				dataType: 'text',
				success: function (data) {
					self.$textarea.val(data);
					self.$title.val(self.model.get('title'));
				},
			});
		},

		save: function(){
			console.log('saving!');
		},

		update: function(){
			this.model.set('content', this.$textarea.val());
		}
	});

	var Interface = Blocks.View.inherits(function(){
		this.list = new List(this);
		this.editor = new Editor();

		this.list.on('select', this.editor.edit, this.editor);

		this.element.appendChild(this.list.element);
		this.element.appendChild(this.editor.element);
	}, {
		element_classes: 'notes_interface'
	});


	return Interface;
});
