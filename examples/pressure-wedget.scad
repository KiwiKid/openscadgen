include <BOSL2/std.scad>;

// Parameters
wedge_length = 70;  // Base length of the wedge
wedge_width = 100;    // Width of the wedge
wedge_height = 25;   // Max height of the wedge
wedge_angle = 15;    // Wedge incline angle
tank_radius = 80;    // Approximate gas tank radius

module pressure_wedge() {
   // difference() {
        // Create the wedge using BOSL2's built-in function
        wedge(size=[wedge_length, wedge_width, wedge_height], anchor=BOTTOM+LEFT+BACK);

        // Subtract a curved section to match tank shape
   //   translate([wedge_length / 2, 0, wedge_height])
    //        rotate([90, 0, 0])
  //          cylinder(r=tank_radius, h=wedge_width, center=true);
  //  }
}

// Render the wedge
pressure_wedge();
