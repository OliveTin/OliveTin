<template>
	<span class="action-icon-glyph">
		<HugeiconsIcon
			v-if="hugeiconsModel"
			:icon="hugeiconsModel"
			width="1em"
			height="1em"
			class="action-icon-glyph-svg"
		/>
		<span v-else v-text="decodedTextGlyph"></span>
	</span>
</template>

<script setup>
import { computed } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { CommandLineIcon } from '@hugeicons/core-free-icons'

const hugeiconsPrefix = 'hugeicons:'

/** Maps config values like hugeicons:CommandLineIcon to Hugeicons icon definitions. */
const hugeiconsRegistry = {
	CommandLineIcon,
}

const props = defineProps({
	glyph: {
		type: String,
		required: false,
		default: '',
	},
})

const hugeiconsModel = computed(() => {
	if (!props.glyph) {
		return CommandLineIcon
	}

	if (!props.glyph.startsWith(hugeiconsPrefix)) {
		return null
	}

	const name = props.glyph.slice(hugeiconsPrefix.length)
	const iconModel = hugeiconsRegistry[name]

	return iconModel ?? CommandLineIcon
})

function decodeHtmlEntities(text) {
	return text.replace(/&#x([0-9a-fA-F]+);?/g, (_, hex) => {
		const codePoint = Number.parseInt(hex, 16)
		return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
	}).replace(/&#(\d+);?/g, (_, decimal) => {
		const codePoint = Number.parseInt(decimal, 10)
		return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
	})
}

const decodedTextGlyph = computed(() => {
	if (hugeiconsModel.value) {
		return ''
	}

	return decodeHtmlEntities(props.glyph)
})
</script>

<style scoped>
.action-icon-glyph {
	display: inline-flex;
	vertical-align: middle;
	align-items: center;
	justify-content: center;
}

.action-icon-glyph-svg {
	vertical-align: middle;
}
</style>
