import os

root_dir = r"c:\Users\t15\Training\GopherShip"
search_dir = os.path.join(root_dir, "_bmad-output")
abs_prefix = "file:///c:/Users/t15/Training/GopherShip/"

for root, dirs, files in os.walk(search_dir):
    for file in files:
        if file.endswith(".md"):
            file_path = os.path.join(root, file)
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()
            
            if abs_prefix in content:
                print(f"Processing {file_path}")
                # Depth of current file relative to root_dir
                rel_base = os.path.relpath(root_dir, root)
                # Ensure it uses forward slashes for markdown compatibility
                rel_base = rel_base.replace(os.sep, "/")
                if rel_base == ".":
                    rel_base = ""
                else:
                    rel_base += "/"
                
                # Simple replacement for prefix
                # Note: this assumes all file:/// links point to the same repo root
                new_content = content.replace(abs_prefix, rel_base)
                
                with open(file_path, "w", encoding="utf-8") as f:
                    f.write(new_content)
