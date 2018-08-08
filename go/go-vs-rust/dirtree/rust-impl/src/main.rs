use std::fs::{self, DirEntry};
use std::io;
use std::path::Path;

fn main() {
    // добавить обработку аргументов командной строки
    // - если знак точки после cargo run, то построение дерева папок в локальной папке
    // - если после точки знак -f, то построение дерева файлов с размером и отсортировано по алфавиту в папках
    let paths = fs::read_dir("./").unwrap();

    for path in paths {
        visit_dirs(&path.unwrap().path(), paths);
        println!("Name: {}", path.unwrap().path().display());
    }
}

// one possible implementation of walking a directory only visiting files
fn visit_dirs(dir: &Path, cb: &Fn(&DirEntry)) -> io::Result<()> {
    if dir.is_dir() {
        for entry in fs::read_dir(dir)? {
            let entry = entry?;
            let path = entry.path();
            if path.is_dir() {
                visit_dirs(&path, cb)?;
            } else {
                cb(&entry);
            }
        }
    }
    Ok(())
}
