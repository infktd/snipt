import { createContext, useContext, useReducer, type Dispatch, type ReactNode } from "react";
import type { AppState, AppAction } from "./types";

const initialState: AppState = {
  snippets: [],
  searchResults: null,
  selectedIds: new Set(),
  anchorId: null,
  focusId: null,
  editMode: false,
  searchQuery: "",
  createMode: false,
  detailView: { kind: "snippet" },
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_SNIPPETS":
      return { ...state, snippets: action.snippets };
    case "SET_SEARCH_RESULTS":
      return { ...state, searchResults: action.results };

    case "SELECT_SINGLE":
      return {
        ...state,
        selectedIds: new Set([action.id]),
        anchorId: action.id,
        focusId: action.id,
        editMode: false,
        createMode: false,
        detailView: { kind: "snippet" },
      };

    case "SELECT_TOGGLE": {
      const next = new Set(state.selectedIds);
      if (next.has(action.id)) {
        next.delete(action.id);
      } else {
        next.add(action.id);
      }
      return {
        ...state,
        selectedIds: next,
        anchorId: action.id,
        focusId: action.id,
        editMode: false,
        createMode: false,
      };
    }

    case "SELECT_RANGE": {
      const { id, list } = action;
      const anchor = state.anchorId;
      if (!anchor) {
        return {
          ...state,
          selectedIds: new Set([id]),
          anchorId: id,
          focusId: id,
          editMode: false,
          createMode: false,
        };
      }
      const anchorIdx = list.indexOf(anchor);
      const targetIdx = list.indexOf(id);
      if (anchorIdx === -1 || targetIdx === -1) {
        return {
          ...state,
          selectedIds: new Set([id]),
          anchorId: id,
          focusId: id,
        };
      }
      const start = Math.min(anchorIdx, targetIdx);
      const end = Math.max(anchorIdx, targetIdx);
      const rangeIds = list.slice(start, end + 1);
      return {
        ...state,
        selectedIds: new Set(rangeIds),
        focusId: id,
        editMode: false,
        createMode: false,
      };
    }

    case "SELECT_CLEAR":
      return {
        ...state,
        selectedIds: new Set(),
        anchorId: null,
        focusId: null,
      };

    case "SET_EDIT_MODE":
      return { ...state, editMode: action.editing };
    case "SET_SEARCH_QUERY":
      return { ...state, searchQuery: action.query };
    case "SET_CREATE_MODE":
      return {
        ...state,
        createMode: action.creating,
        editMode: action.creating,
        selectedIds: new Set(),
        anchorId: null,
        focusId: null,
        detailView: { kind: "snippet" },
      };
    case "CLEAR_SEARCH":
      return { ...state, searchQuery: "", searchResults: null };
    case "OPEN_SETTINGS":
      return {
        ...state,
        detailView: { kind: "settings" },
        editMode: false,
        createMode: false,
      };
    case "CLOSE_SETTINGS":
      return {
        ...state,
        detailView: { kind: "snippet" },
        createMode: false,
      };
    default:
      return state;
  }
}

const AppStateContext = createContext<AppState>(initialState);
const AppDispatchContext = createContext<Dispatch<AppAction>>(() => {});

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState);
  return (
    <AppStateContext.Provider value={state}>
      <AppDispatchContext.Provider value={dispatch}>
        {children}
      </AppDispatchContext.Provider>
    </AppStateContext.Provider>
  );
}

export function useAppState() {
  return useContext(AppStateContext);
}

export function useAppDispatch() {
  return useContext(AppDispatchContext);
}
