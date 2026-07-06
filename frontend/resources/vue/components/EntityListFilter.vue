<template>
	<label class="input-with-icons">
		<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
			<path fill="currentColor"
				d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
		</svg>
		<input
			:value="modelValue"
			placeholder="Filter entities..."
			@input="$emit('update:modelValue', $event.target.value)"
		/>
		<button title="Clear search filter" :disabled="!modelValue" @click="$emit('update:modelValue', '')">
			<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
				<path fill="currentColor"
					d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
			</svg>
		</button>
	</label>
</template>

<script setup>
	defineProps({
		modelValue: {
			type: String,
			default: ''
		}
	})

	defineEmits(['update:modelValue'])
</script>

<style scoped>
.input-with-icons {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem;
	border: 1px solid var(--border-color, #ccc);
	border-radius: 0.25rem;
	background: var(--section-background);
	width: 100%;
	max-width: 300px;
}

.input-with-icons input {
	border: none;
	outline: none;
	background: transparent;
	flex: 1;
	color: var(--text-primary);
}

.input-with-icons button {
	background: none;
	border: none;
	cursor: pointer;
	color: var(--text-secondary);
}

.input-with-icons button:disabled {
	opacity: 0.3;
	cursor: not-allowed;
}
</style>
