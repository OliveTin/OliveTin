<template>
	<Section 
		title="Table Example" 
		subtitle="A simple table with sortable columns and pagination"
		:padding = "false"
		>
		 <template #toolbar>
			 <button>Add Item</button>
		 </template>

		<Table 
			 :data = "data" 
			 :headers = "headers" 
			 :show-pagination = "true"
			 >
			 <template #cell-name="{ row, value }">
				 <router-link :to="{ name: 'ViewItem', params: { id: row.name } }">
					 {{ value }}
				 </router-link>
			 </template>

			 <template #cell-city="{ row, value }">
				 <span class = "subtle">
				 {{ value }}
				 </span>
			 </template>
		</Table>
	</Section>
</template>

<script setup>
	import { ref, onMounted } from 'vue'

	const headers = ref([
		{ key: "name", label: "Name", sortable: true, hidden: false, width: "20%" },
		{ key: "age", label: "Age", sortable: false },
		{ key: "city", label: "City", sortable: true, width: "200px" },
	])

	const data = ref([])

	onMounted(() => {
		data.value = [
			{ name: "Alice", age: 30, city: "New York", foo: "bar" },
			{ name: "Bob", age: 25, city: "Los Angeles" },
			{ name: "Charlie", age: 35, city: "Chicago" },
			{ name: "David", age: 28, city: "Houston" },
			{ name: "Eve", age: 22, city: "Phoenix" },
			{ name: "Frank", age: 40, city: "Philadelphia" },
			{ name: "Grace", age: 29, city: "San Antonio" },
			{ name: "Hannah", age: 32, city: "San Diego" },
			{ name: "Ian", age: 27, city: "Dallas" },
			{ name: "Jack", age: 31, city: "San Jose" },
			{ name: "Kathy", age: 26, city: "Austin" },
			{ name: "Leo", age: 33, city: "Jacksonville" },
			{ name: "Mia", age: 24, city: "Fort Worth" },
			{ name: "Nina", age: 38, city: "Columbus" },
			{ name: "Oscar", age: 36, city: "Charlotte" },
			{ name: "Paul", age: 34, city: "San Francisco" },
			{ name: "Quinn", age: 23, city: "Indianapolis" },
			{ name: "Rita", age: 37, city: "Seattle" },
			{ name: "Sam", age: 39, city: "Denver" },
			{ name: "Tina", age: 21, city: "Washington D.C." },
			{ name: "Uma", age: 28, city: "Miami" },
			{ name: "Victor", age: 30, city: "Portland" },
			{ name: "Wendy", age: 29, city: "Nashville" },
			{ name: "Xavier", age: 31, city: "Oklahoma City" },
			{ name: "Yara", age: 27, city: "Tampa" },
			{ name: "Zane", age: 32, city: "Albuquerque" },
			{ name: "Ava", age: 26, city: "Las Vegas" },
			{ name: "Ben", age: 33, city: "New Orleans" },
			{ name: "Cara", age: 25, city: "San Francisco" },
			{ name: "Drew", age: 34, city: "Seattle" },
			{ name: "Ethan", age: 28, city: "Denver" },
			{ name: "Fiona", age: 30, city: "Washington D.C." },
			{ name: "George", age: 29, city: "Miami" },
			{ name: "Hannah", age: 32, city: "San Diego" },
			{ name: "Ian", age: 27, city: "Dallas" },
			{ name: "Jack", age: 31, city: "San Jose" },
			{ name: "Kathy", age: 26, city: "Austin" },
			{ name: "Leo", age: 33, city: "Jacksonville" },
			{ name: "Mia", age: 24, city: "Fort Worth" },
			{ name: "Nina", age: 38, city: "Columbus" },
			{ name: "Oscar", age: 36, city: "Charlotte" },
			{ name: "Paul", age: 34, city: "San Francisco" },
			{ name: "Quinn", age: 23, city: "Indianapolis" },
			{ name: "Rita", age: 37, city: "Seattle" },
			{ name: "Sam", age: 39, city: "Denver" },
			{ name: "Tina", age: 21, city: "Washington D.C." },
			{ name: "Uma", age: 28, city: "Miami" },
			{ name: "Victor", age: 30, city: "Portland" },
			{ name: "Wendy", age: 29, city: "Nashville" },
			{ name: "Xavier", age: 31, city: "Oklahoma City" },
			{ name: "Yara", age: 27, city: "Tampa" },
			{ name: "Zane", age: 32, city: "Albuquerque" },
			{ name: "Ava", age: 26, city: "Las Vegas" },
			{ name: "Ben", age: 33, city: "New Orleans" },
			{ name: "Cara", age: 25, city: "San Francisco" },
			{ name: "Drew", age: 34, city: "Seattle" },
			{ name: "Ethan", age: 28, city: "Denver" },
			{ name: "Fiona", age: 30, city: "Washington D.C." },
			{ name: "George", age: 29, city: "Miami" },
			{ name: "Hannah", age: 32, city: "San Diego" },
			{ name: "Ian", age: 27, city: "Dallas" },
			{ name: "Jack", age: 31, city: "San Jose" },
			{ name: "Kathy", age: 26, city: "Austin" },
			{ name: "Leo", age: 33, city: "Jacksonville" },
			{ name: "Mia", age: 24, city: "Fort Worth" },
			{ name: "Nina", age: 38, city: "Columbus" },
			{ name: "Oscar", age: 36, city: "Charlotte" },
			{ name: "Paul", age: 34, city: "San Francisco" },
			{ name: "Quinn", age: 23, city: "Indianapolis" },
			{ name: "Rita", age: 37, city: "Seattle" },
			{ name: "Sam", age: 39, city: "Denver" },
			{ name: "Tina", age: 21, city: "Washington D.C." },
			{ name: "Uma", age: 28, city: "Miami" },
			{ name: "Victor", age: 30, city: "Portland" },
		]
	})
</script>
