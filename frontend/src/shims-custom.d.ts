// Custom ambient module declarations to silence TypeScript for certain imports
declare module '@/utils/formValidate'
declare module '@/api/merchant/index'

// Allow importing .vue single-file components without errors
declare module '*.vue' {
  import { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// Allow importing static assets directly
declare module '*.png'
declare module '*.jpg'
declare module '*.jpeg'
declare module '*.svg'
declare module '*.gif'

// Third-party libraries used without @types
declare module 'vue-router'
declare module 'element-plus'
