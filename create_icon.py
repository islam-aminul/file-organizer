#!/usr/bin/env python3
"""
Simple script to create a basic ZenSort icon
Requires Pillow: pip install Pillow
"""

from PIL import Image, ImageDraw, ImageFont
import os

def create_icon():
    # Create a 256x256 image with transparent background
    size = 256
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Draw a folder-like shape with gradient colors
    # Main folder body
    folder_color = (52, 152, 219)  # Blue
    accent_color = (41, 128, 185)  # Darker blue
    
    # Folder tab
    tab_points = [(40, 80), (100, 80), (110, 60), (160, 60), (170, 80), (216, 80), (216, 200), (40, 200)]
    draw.polygon(tab_points, fill=folder_color, outline=accent_color, width=3)
    
    # Add organizing arrows/lines to represent file sorting
    arrow_color = (255, 255, 255, 200)  # Semi-transparent white
    
    # Draw sorting lines
    for i, y in enumerate([110, 130, 150, 170]):
        width = 120 - (i * 10)
        x_start = 60
        draw.rectangle([x_start, y, x_start + width, y + 4], fill=arrow_color)
        
        # Add small arrow at the end
        arrow_x = x_start + width
        arrow_points = [(arrow_x, y), (arrow_x + 8, y + 2), (arrow_x, y + 4)]
        draw.polygon(arrow_points, fill=arrow_color)
    
    # Add "Z" letter for ZenSort
    try:
        # Try to use a system font
        font_size = 48
        font = ImageFont.truetype("arial.ttf", font_size)
    except:
        # Fallback to default font
        font = ImageFont.load_default()
    
    # Draw "Z" in the top-right corner
    text_color = (255, 255, 255)
    draw.text((180, 30), "Z", fill=text_color, font=font)
    
    # Save as ICO file with multiple sizes
    icon_sizes = [(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
    
    # Create images for each size
    images = []
    for icon_size in icon_sizes:
        resized = img.resize(icon_size, Image.Resampling.LANCZOS)
        images.append(resized)
    
    # Save as ICO
    img.save('icon.ico', format='ICO', sizes=[(img.width, img.height) for img in images])
    print("Icon created successfully: icon.ico")

if __name__ == "__main__":
    create_icon()
