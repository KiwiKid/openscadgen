/**
   mulit-size upholstery screw-in button
   
   Source: https://github.com/KiwiKid/openscaden/
   Author: KiwiKid
   
*/

$fn = 100;



// Size of the screw hole
screw_hole_radius_size = 2.5;
// Size of gap for screw head
screw_head_hole_radius_size = 5.3; //max(screw_hole_radius_size*1.5, 5);

// The gap between the bottom of the button and the screw hole
screw_hole_guard_depth = 1;

// Controls the 'height' of the button
button_height =5;
// Controls the amount of 'curve' in the button
button_radius = 90;

// Control how thick/proud the button is
button_width = 3;

cube_side_length = button_radius*4;
cube_size = [cube_side_length, cube_side_length, cube_side_length];

all_the_way = 1000;

optional_part_id_letter = "";


// Will be populated by openscadgen
module label_text(txt) {
    if (!is_undef(txt)) {
     
        text_size = 3;  // Size of the text
        text_height = 3; // Height of the extruded text
        text_y_translate = screw_head_hole_radius_size+1;
        text_depth = 0.8;
        
        // Estimate width of each character
        char_width = text_size * 0.94;

        // Get the number of characters in the text
        total_width = len(txt) * char_width;

        // Center the text on the X-axis by adjusting the translate value
        translate([-total_width / 2, text_y_translate, -button_width+text_depth])
        rotate([180, 0, 0])
        linear_extrude(height =text_height)
        text(txt, size = text_size, font = "Liberation Sans:style=Bold");
    }
}

difference(){

    translate([0, 0, -button_radius-5+button_height])
    sphere(button_radius);
        
    translate([-cube_side_length+button_radius, -button_radius, -cube_size[2]-button_width])
    cube(cube_size);
    
    
    translate([0, 0, -100])
    cylinder(h=all_the_way, r = screw_hole_radius_size);
    
    translate([0, 0, screw_hole_guard_depth-button_width])
    cylinder(h=all_the_way, r = screw_head_hole_radius_size);
    
    
    
  
    // (optional) part label to help with identification
    label_text(optional_part_id_letter);


}
