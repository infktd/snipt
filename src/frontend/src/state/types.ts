export interface Snippet {
  ID: string;
  Title: string;
  Content: string;
  Language: string;
  Description: string;
  Source: string;
  Pinned: boolean;
  UseCount: number;
  Tags: string[];
  CreatedAt: string;
  UpdatedAt: string;
}

export interface SearchResult {
  Snippet: Snippet;
  Score: number;
  TitleIndices: number[] | null;
}

export interface ListOpts {
  Language?: string;
  Tag?: string;
  Pinned?: boolean;
  Sort?: string;
}

export interface Stats {
  TotalSnippets: number;
  TotalTags: number;
  Languages: Record<string, number>;
  MostUsed: Snippet | null;
  RecentlyAdded: Snippet[];
}

export interface AppState {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  selectedId: string | null;
  editMode: boolean;
  searchQuery: string;
  createMode: boolean;
}

export type AppAction =
  | { type: "SET_SNIPPETS"; snippets: Snippet[] }
  | { type: "SET_SEARCH_RESULTS"; results: SearchResult[] | null }
  | { type: "SET_SELECTED"; id: string | null }
  | { type: "SET_EDIT_MODE"; editing: boolean }
  | { type: "SET_SEARCH_QUERY"; query: string }
  | { type: "SET_CREATE_MODE"; creating: boolean }
  | { type: "CLEAR_SEARCH" };
