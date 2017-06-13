import worldManger from "../plugins/worldmanager"

import jsonApi from "../plugins/jsonapi"


export default {
  subTableColumns(state) {
    return state.subTableColumns
  },
  isAuthenticated (state) {
    // console.log("check is authenticated: ", window.localStorage.getItem("token"))
    return !!window.localStorage.getItem("token")
  },
  systemActions(state) {
    return state.systemActions;
  },
  authToken (state) {
    return window.localStorage.getItem("token")
  },
  selectedAction (state) {
    return state.selectedAction;
  },
  selectedInstanceReferenceId (state) {
    return state.selectedInstanceReferenceId
  },
  user (state) {
    return JSON.parse(window.localStorage.getItem("user"))
  },
  actions (state) {
    return state.actions;
  },
  selectedTable (state) {
    console.log("get selected table", state.selectedTable)
    return state.selectedTable;
  },
  finder (state) {
    return state.finder
  },
  selectedRow (state) {
    return state.selectedRow;
  },
  selectedTableColumns (state) {
    return state.selectedTableColumns;
  },
  selectedInstanceReferenceId(state) {
    return state.selectedInstanceReferenceId
  },
  selectedSubTable (state) {
    return state.selectedSubTable
  },
  showAddEdit (state) {
    return state.showAddEdit;
  },
  visibleWorlds (state) {
    let filtered = worldManger.getWorlds().filter(function (w, r) {
      if (!state.selectedInstanceReferenceId) {
        // console.log("No selected item. Return top level tables")
        return w.is_top_level == '1' && w.is_hidden == '0';
      } else {
        console.log("Selected item found. Return child tables")
        const model = jsonApi.modelFor(w.table_name);
        const attrs = model["attributes"];
        const keys = Object.keys(attrs);
        if (keys.indexOf(state.selectedTable + "_id") > -1) {
          return w.is_top_level == '0';
        }
        return false;
      }
    });
    console.log("filtered worlds: ", filtered)

    return filtered;
  },
  topWorlds (state) {
    return worldManger.getWorlds().filter(function (w, r) {
      return w.is_top_level == '1' && w.is_hidden == '0';
    });
  }
}
