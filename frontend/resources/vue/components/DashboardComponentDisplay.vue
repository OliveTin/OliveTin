<template>
    <div class="display">
        <div v-html="component.title" />
    </div>
</template>

<script setup>
const props = defineProps({
    component: {
        type: Object,
        required: true
    }
})
</script>

<style scoped>
.display {
	padding: 1em;
	border-radius: .7em;
	box-shadow: 0 0 .6em #aaa;
	text-align: center;
	font-size: small;
	display: flex;
	flex-direction: column;
	flex-grow: 1;
	justify-content: center;
	align-items: center;
}

@media (prefers-color-scheme: dark) {
    .display {
        border-color: #000;
        box-shadow: 0 0 .6em #000;
    }
}

</style>