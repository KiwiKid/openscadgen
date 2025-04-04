include <BOSL2/std.scad>  // BOSL2 library inclusion

// Parameters for cylindrical path and text
text = "Steven Cromb";
font_size = 12;
pi = 3.14;

// Create a cylindrical extrude using path_text
module cylindrical_extruded_text() {
    // Create the path text
    path = reverse(path3d(arc(100, r=150, angle=[65, 190])));

    #path_text(text = text, font = "Arial", size = font_size, path=path, textmetrics=true, thickness=10);
    
}

// Render the cylindrical text
cylindrical_extruded_text();
