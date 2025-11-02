// Vuex store (简单实现，也可以使用 Pinia)
import { createStore } from 'vuex'
import { getUser, getPermissions, getToken } from '@/utils/auth'

const store = createStore({
  state: {
    user: null,
    permissions: [],
    token: null,
    isLoggedIn: false
  },
  
  mutations: {
    SET_USER(state, user) {
      state.user = user
      state.isLoggedIn = !!user
    },
    
    SET_PERMISSIONS(state, permissions) {
      state.permissions = permissions
    },
    
    SET_TOKEN(state, token) {
      state.token = token
    },
    
    CLEAR_AUTH(state) {
      state.user = null
      state.permissions = []
      state.token = null
      state.isLoggedIn = false
    }
  },
  
  actions: {
    initAuth({ commit }) {
      const user = getUser()
      const permissions = getPermissions()
      const token = getToken()
      
      if (user && token) {
        commit('SET_USER', user)
        commit('SET_PERMISSIONS', permissions)
        commit('SET_TOKEN', token)
      }
    },
    
    clearAuth({ commit }) {
      commit('CLEAR_AUTH')
    }
  },
  
  getters: {
    isLoggedIn: state => state.isLoggedIn,
    user: state => state.user,
    permissions: state => state.permissions,
    hasPermission: state => permissionCode => {
      return state.permissions.some(p => p.permission_code === permissionCode)
    }
  }
})

export default store