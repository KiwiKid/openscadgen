$fn = 200;

// Attempt to close paths using offset
//difference(mode=hacksaw);

scale([10, 10, 10])
translate([-4.05, -4.25, 0.5])
linear_extrude(height=1, center=true)
    import("./example.svg");

//    translate([4, 4.1, size])
scale([10, 10, 10])

            cylinder(r=5, h=1, center=true);