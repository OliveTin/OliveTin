<template>
	<section :class = "'with-header-and-content ' + classes" :id = "id">
		<div class = "section-header flex-row">
			<div class = "fg1">
				<h2 :title = "subtitle">{{ title }}</h2>
			</div>
			<div role = "toolbar">
				<slot name="toolbar" />
			</div>
		</div>
		<div :class = "padding ? 'section-content padding' : 'section-content'">
			<slot />
		</div>
	</section>
</template>

<script setup>

const props = defineProps({
	title: {
		type: String,
		default: "Untitled Section",
	},
	padding: {
		type: Boolean,
		default: true,
	},
	subtitle: {
		type: String,
		default: "",
	},
	id: {
		type: String,
		default: "",
	},
	classes: {
		type: String,
		default: "",
	},
});
</script>

<style scoped>
:slotted(.section-header button).neutral,
:slotted(.section-header .button).neutral
{
	background-color: white;
	color: #000;
}

:slotted(.section-header .button:hover).neutral,
:slotted(.section-header .button:hover).neutral
{
	background-color: #f0f0f0;
}

@media (prefers-color-scheme: dark) {
	:slotted(.section-header button).neutral,
	:slotted(.section-header .button).neutral
	{
		background-color: #111;
		border: 1px solid #222;
		color: #fff;
	}

	:slotted(.section-header button:hover).neutral,
	:slotted(.section-header .button:hover).neutral
	{
		background-color: #222;
		border: 1px solid #333;
		color: #fff;
	}
}
</style>