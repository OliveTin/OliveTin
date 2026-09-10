import neostandard from 'neostandard'
import pluginVue from 'eslint-plugin-vue'

export default [
	{
		ignores: [
			'dist/**',
			'node_modules/**',
			'resources/scripts/gen/**'
		]
	},
	...neostandard({
		env: ['browser'],
		noJsx: true
	}),
	...pluginVue.configs['flat/recommended'],
	{
		files: ['**/*.{js,mjs,vue}'],
		languageOptions: {
			ecmaVersion: 'latest',
			sourceType: 'module'
		},
		rules: {
			'@stylistic/no-tabs': 'off',
			'@stylistic/no-mixed-spaces-and-tabs': 'off',
			'vue/multi-word-component-names': 'off',
			'vue/require-default-prop': 'off',
			'vue/no-v-html': 'warn'
		}
	}
]
