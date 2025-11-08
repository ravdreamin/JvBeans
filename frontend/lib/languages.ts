export function getLanguageFromExtension(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase();

  const languageMap: Record<string, string> = {
    'js': 'javascript',
    'jsx': 'javascript',
    'ts': 'typescript',
    'tsx': 'typescript',
    'py': 'python',
    'java': 'java',
    'c': 'c',
    'cpp': 'cpp',
    'cc': 'cpp',
    'cxx': 'cpp',
    'go': 'go',
    'rs': 'rust',
    'rb': 'ruby',
    'php': 'php',
    'swift': 'swift',
    'kt': 'kotlin',
    'cs': 'csharp',
  };

  return ext && languageMap[ext] ? languageMap[ext] : 'javascript';
}

export function getLanguageLabel(language: string): string {
  const labelMap: Record<string, string> = {
    'javascript': 'JavaScript',
    'typescript': 'TypeScript',
    'python': 'Python',
    'java': 'Java',
    'c': 'C',
    'cpp': 'C++',
    'go': 'Go',
    'rust': 'Rust',
    'ruby': 'Ruby',
    'php': 'PHP',
    'swift': 'Swift',
    'kotlin': 'Kotlin',
    'csharp': 'C#',
  };

  return labelMap[language] || language;
}

export function getLanguageColor(language: string): string {
  const colorMap: Record<string, string> = {
    'javascript': '#5B8CFF',
    'typescript': '#44D6FF',
    'python': '#22E6B8',
    'java': '#30E394',
    'c': '#5B8CFF',
    'cpp': '#44D6FF',
    'go': '#22E6B8',
    'rust': '#30E394',
  };

  return colorMap[language] || '#5B8CFF';
}
