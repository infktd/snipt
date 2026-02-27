import { createContext, useContext, useReducer, type Dispatch, type ReactNode } from "react";
import type { AppState, AppAction } from "./types";

const initialState: AppState = {
  snippets: [],
  searchResults: null,
  selectedId: null,
  editMode: false,
  searchQuery: "",
  createMode: false,
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_SNIPPETS":
      return { ...state, snippets: action.snippets };
    case "SET_SEARCH_RESULTS":
      return { ...state, searchResults: action.results };
    case "SET_SELECTED":
      return { ...state, selectedId: action.id, editMode: false, createMode: false };
    case "SET_EDIT_MODE":
      return { ...state, editMode: action.editing };
    case "SET_SEARCH_QUERY":
      return { ...state, searchQuery: action.query };
    case "SET_CREATE_MODE":
      return { ...state, createMode: action.creating, editMode: true, selectedId: null };
    case "CLEAR_SEARCH":
      return { ...state, searchQuery: "", searchResults: null };
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
