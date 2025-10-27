<template>
	<Section title = "Welcome" classes = "">
		<p>
			Welcome to Femtocrank, a web-based tool for analyzing and visualizing the performance of your code.
			To get started, you can either upload a file or paste your code directly into the editor.
		</p>
		<p>
			For more information on how to use Femtocrank, please refer to the <a href = "/docs">documentation</a>.
		</p>
	</Section>

	<Section title = "Lorum" id = "lorem">
		<template #toolbar>
			<button>
				<HugeiconsIcon :icon = "HighlighterIcon" width = "1em" height = "1em" />
				Highlight</button>
			<button>Quote</button>
			<button class = "good">Add</button>
		</template>

		<div>
			<button class = "neutral">
				Neutral
			</button>
			<button class = "good">
				Good
			</button>
			<button class = "bad">
				Bad
			</button>
		</div>

		<p>
			Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
			Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
			Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
			Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
		</p>
	</Section>

	<Section title = "Shakespeare" id = "shakespeare">
		<p>
			To be, or not to be, that is the question:
			Whether 'tis nobler in the mind to suffer
			The slings and arrows of outrageous fortune,
			Or to take arms against a sea of troubles
			And by opposing end them. To die: to sleep;
			No more; and by a sleep to say we end
			The heart-ache and the thousand natural shocks
			That flesh is heir to, 'tis a consummation
			Devoutly to be wish'd. To die, to sleep;
			To sleep: perchance to dream: ay, there's the rub;
			For in that sleep of death what dreams may come
			When we have shuffled off this mortal coil,
			Must give us pause: there's the respect
			That makes calamity of so long life;
			For who would bear the whips and scorns of time,
			The oppressor's wrong
		</p>
	</Section>

	<Section title = "More Lorem" id = "more-lorem">
		<p>
			Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
			Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
			Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
			Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
		</p>
	</Section>

</template>

<script setup>
	import Section from '../components/Section.vue';
	import { HugeiconsIcon } from '@hugeicons/vue';
	import { HighlighterIcon } from '@hugeicons/core-free-icons';
</script>
